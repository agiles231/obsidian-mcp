package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type WriteNote struct {
	registry *vault.Registry
}

func NewWriteNote(r *vault.Registry) *WriteNote {
	return &WriteNote{registry: r}
}

func (t *WriteNote) Name() string { return "write_note" }

func (t *WriteNote) Description() string {
	return "Create or overwrite a note in the vault. Parent directories are created if needed"
}

func (t *WriteNote) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"ref": {
				Type:        "string",
				Description: "Note reference: a URN (urn:obsidian::vault:note:path/to/note.md) or base path",
			},
			"content": {
				Type:        "string",
				Description: "The full content to write to the note",
			},
		},
		Required: []string{"ref", "content"},
	}
}

type writeNoteArgs struct {
	Ref     string `json:"ref"`
	Content string `json:"content"`
}

func (t *WriteNote) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a writeNoteArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}

	ref, err := urn.ParseRef(a.Ref, t.registry.DefaultName())
	if err != nil {
		return nil, err
	}

	v, err := t.registry.Get(ref.Vault)
	if err != nil {
		return nil, err
	}

	if err := v.WriteFile(ctx, ref.Path, []byte(a.Content)); err != nil {
		return nil, err
	}
	return []mcp.Content{mcp.Text("write " + ref.URN())}, nil
}
