package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type ListObjects struct {
	registry *vault.Registry
}

func NewListObjects(r *vault.Registry) *ListObjects {
	return &ListObjects{registry: r}
}

func (t *ListObjects) Name() string {
	return "list_objects"
}

func (t *ListObjects) Description() string {
	return "List objects (notes, folders, attachments, canvases) in a vault. Returns URN, type, ane name for each."
}

func (t *ListObjects) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"path": {
				Type:        "string",
				Description: "Folder path relative to the vault root. Empty string lists the root.",
			},
			"types": {
				Type:        "array",
				Description: "Filter by type: note, folder, attachment, canvas. Empty = all.",
			},
			"recursive": {
				Type:        "boolean",
				Description: "Include subdirectories. Default false.",
			},
		},
		Required: []string{},
	}
}

type listObjectsArgs struct {
	Path      string   `json:"path"`
	Types     []string `json:"types"`
	Recursive bool     `json:"recursive"`
}

type objectResult struct {
	URN  string `json:"urn"`
	Type string `json:"type"`
	Name string `json:"name"`
}

func (t *ListObjects) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a listObjectsArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}

	v, err := t.registry.Default()
	if err != nil {
		return nil, err
	}

	// Build type filter
	var typeFilter map[string]bool
	if len(a.Types) > 0 {
		typeFilter = make(map[string]bool, len(a.Types))
		for _, t := range a.Types {
			typeFilter[t] = true
		}
	}

	entries, err := v.ListObjects(ctx, a.Path, vault.ListOptions{
		Types:     typeFilter,
		Recursive: a.Recursive,
	})
	if err != nil {
		return nil, err
	}

	// Convert to output format
	results := make([]objectResult, len(entries))
	for i, e := range entries {
		ref := urn.Ref{Vault: v.Name(), Type: e.Type, Path: e.Path}
		results[i] = objectResult{
			URN:  ref.URN(),
			Type: e.Type,
			Name: e.Name,
		}
	}

	data, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return []mcp.Content{mcp.Text(string(data))}, nil
}

func (t *ListObjects) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:         "List Objects",
		ReadOnlyHint:  mcp.HintTrue(),
		OpenWorldHint: mcp.HintFalse(),
	}
}
