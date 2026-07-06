package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type ReadNote struct {
	registry *vault.Registry
}

func NewReadNote(r *vault.Registry) *ReadNote {
	return &ReadNote{registry: r}
}

func (t *ReadNote) Name() string { return "read_note" }

func (t *ReadNote) Description() string {
	return "Read the contents of a note from the vault. Accepts a URN (urn:obsidian::vault:note:path)"
}

func (t *ReadNote) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"ref": {
				Type:        "string",
				Description: "Note references: a URN (urn:obsidian::vault:note:path/to/note.md)",
			},
		},
		Required: []string{"ref"},
	}
}

type readNoteArgs struct {
	Ref string `json:"ref"`
}

func (t *ReadNote) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a readNoteArgs
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

func (t *ReadNote) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:         "Read Note",
		ReadOnlyHint:  mcp.HintTrue(),
		OpenWorldHint: mcp.HintFalse(),
	}
}
