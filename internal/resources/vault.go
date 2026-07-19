// Package resources implements MCP resources over vault objects.
//
// Resource URIs are canonical urn:obsidian: URNs (see docs/urn-spec.md).
// resources/list pages readable file objects (notes, canvases, attachments);
// folders are discovery-only via list_objects and are not listed as resources.
// resources/read returns text for notes/canvases and base64 blobs for other
// attachments, subject to the same allow/deny rules as tools.
package resources

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

const defaultPageSize = 100

// VaultResources backs MCP resources/list and resources/read from a vault registry.
type VaultResources struct {
	registry *vault.Registry
	pageSize int
}

// NewVaultResources creates a Resources implementation for the given registry.
func NewVaultResources(r *vault.Registry) *VaultResources {
	return &VaultResources{registry: r, pageSize: defaultPageSize}
}

// List returns a page of readable vault file objects as resource descriptors.
// Cursor is a decimal offset string (empty = start).
func (r *VaultResources) List(ctx context.Context, cursor string) ([]mcp.ResourceDescriptor, string, error) {
	v, err := r.registry.Default()
	if err != nil {
		return nil, "", err
	}

	offset, err := parseCursor(cursor)
	if err != nil {
		return nil, "", fmt.Errorf("invalid cursor: %w", err)
	}

	entries, err := v.ListObjects(ctx, "", vault.ListOptions{
		Types: map[string]bool{
			urn.TypeNote:       true,
			urn.TypeCanvas:     true,
			urn.TypeAttachment: true,
		},
		Recursive: true,
	})
	if err != nil {
		return nil, "", err
	}

	if offset > len(entries) {
		offset = len(entries)
	}
	end := offset + r.pageSize
	if end > len(entries) {
		end = len(entries)
	}
	page := entries[offset:end]

	out := make([]mcp.ResourceDescriptor, 0, len(page))
	for _, e := range page {
		ref := urn.Ref{Vault: v.Name(), Type: e.Type, Path: e.Path}
		uri := ref.URN()
		out = append(out, mcp.ResourceDescriptor{
			URI:         uri,
			Name:        e.Name,
			Title:       e.Path,
			Description: resourceDescription(e.Type, e.Path),
			MimeType:    mimeForType(e.Type, e.Path),
		})
	}

	var next string
	if end < len(entries) {
		next = strconv.Itoa(end)
	}
	return out, next, nil
}

// Read returns the contents of the resource identified by uri (a canonical URN
// or bare vault-relative path accepted by urn.ParseRef).
func (r *VaultResources) Read(ctx context.Context, uri string) ([]mcp.ResourceContents, error) {
	ref, err := urn.ParseRef(uri, r.registry.DefaultName())
	if err != nil {
		return nil, mcp.ErrResourceNotFound
	}
	// Folders have no file body as a resource.
	if ref.Type == urn.TypeFolder {
		return nil, mcp.ErrResourceNotFound
	}

	v, err := r.registry.Get(ref.Vault)
	if err != nil {
		return nil, mcp.ErrResourceNotFound
	}

	data, err := v.ReadFile(ctx, ref.Path)
	if err != nil {
		// Map access failures to not-found (ADR-0005: deny looks like missing).
		return nil, mcp.ErrResourceNotFound
	}

	// Canonical URI in the response even if the client sent a bare path.
	outURI := ref.URN()
	mime := mimeForType(ref.Type, ref.Path)

	if isTextResource(ref.Type, ref.Path, data) {
		return []mcp.ResourceContents{
			mcp.TextResource(outURI, mime, string(data)),
		}, nil
	}
	return []mcp.ResourceContents{
		mcp.BlobResource(outURI, mime, data),
	}, nil
}

// ListTemplates advertises the URN template for direct resource access.
func (r *VaultResources) ListTemplates(_ context.Context, cursor string) ([]mcp.ResourceTemplateDescriptor, string, error) {
	if cursor != "" {
		return nil, "", nil
	}
	return []mcp.ResourceTemplateDescriptor{{
		URITemplate: "urn:obsidian::{vault}:{type}:{+path}",
		Name:        "Vault object",
		Title:       "Obsidian vault object",
		Description: "Read any vault file by canonical URN. type is note, canvas, or attachment; path is vault-relative.",
		MimeType:    "text/markdown",
	}}, "", nil
}

func parseCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(cursor)
	if err != nil || n < 0 {
		return 0, errors.New("cursor must be a non-negative integer")
	}
	return n, nil
}

func resourceDescription(objType, relPath string) string {
	switch objType {
	case urn.TypeNote:
		return "Markdown note: " + relPath
	case urn.TypeCanvas:
		return "Obsidian canvas: " + relPath
	case urn.TypeAttachment:
		return "Attachment: " + relPath
	default:
		return relPath
	}
}

func mimeForType(objType, relPath string) string {
	switch objType {
	case urn.TypeNote:
		return "text/markdown"
	case urn.TypeCanvas:
		return "application/json"
	case urn.TypeAttachment:
		return mimeFromExt(path.Ext(relPath))
	default:
		return "application/octet-stream"
	}
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "text/javascript"
	case ".csv":
		return "text/csv"
	case ".xml":
		return "application/xml"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".canvas":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// isTextResource reports whether content should be returned as text (vs base64 blob).
func isTextResource(objType, relPath string, data []byte) bool {
	switch objType {
	case urn.TypeNote, urn.TypeCanvas:
		return true
	}
	mime := mimeFromExt(path.Ext(relPath))
	// Known binary families always use blob, even if bytes happen to be UTF-8.
	if mime == "application/octet-stream" || mime == "application/pdf" ||
		(strings.HasPrefix(mime, "image/") && mime != "image/svg+xml") {
		return false
	}
	if !utf8.Valid(data) {
		return false
	}
	if strings.IndexByte(string(data), 0) >= 0 {
		return false
	}
	return true
}

