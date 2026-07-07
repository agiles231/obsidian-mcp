package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type ReadFile struct {
	registry *vault.Registry
}

func NewReadFile(r *vault.Registry) *ReadFile {
	return &ReadFile{registry: r}
}

func (t *ReadFile) Name() string { return "read_file" }

func (t *ReadFile) Description() string {
	return "Read the contents of a file from the vault. Accepts a URN (urn:obsidian::vault:note:path)"
}

func (t *ReadFile) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"ref": {
				Type:        "string",
				Description: "File references: a URN (urn:obsidian::vault:note:path/to/note.md)",
			},
		},
		Required: []string{"ref"},
	}
}

type readFileArgs struct {
	Ref string `json:"ref"`
}

func (t *ReadFile) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a readFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}

	// Parse the reference (URN or bare path)
	ref, err := urn.ParseRef(a.Ref, t.registry.DefaultName())
	if err != nil {
		return nil, err
	}

	// Look up the vault
	v, err := t.registry.Get(ref.Vault)
	if err != nil {
		return nil, err
	}

	// Read the file
	data, err := v.ReadFile(ctx, ref.Path)
	if err != nil {
		return nil, err
	}

	// TODO: if the ref.Anchor is set, extract just that section
	return []mcp.Content{mcp.Text(string(data))}, nil
}

func (t *ReadFile) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:         "Read File",
		ReadOnlyHint:  mcp.HintTrue(),
		OpenWorldHint: mcp.HintFalse(),
	}
}
