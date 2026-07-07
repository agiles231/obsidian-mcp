package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type AppendNote struct {
	registry *vault.Registry
}

func NewAppendNote(r *vault.Registry) *AppendNote {
	return &AppendNote{registry: r}
}

func (t *AppendNote) Name() string { return "append_note" }

func (t *AppendNote) Description() string {
	return "Append content to a note. Creates the note if it doesn't exist. Content is appended as-is; include leading newlines for separation if needed."
}

func (t *AppendNote) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"ref": {
				Type:        "string",
				Description: "Note reference: a URN (urn:obsdian::vault:note:path/to/note.md) or base path.",
			},
			"content": {
				Type:        "string",
				Description: "Content to append to the note",
			},
		},
		Required: []string{"ref", "content"},
	}
}

type appendNoteArgs struct {
	Ref     string `json:"ref"`
	Content string `json:"content"`
}

func (t *AppendNote) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a appendNoteArgs
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

	if err := v.AppendFile(ctx, ref.Path, []byte(a.Content)); err != nil {
		return nil, err
	}

	return []mcp.Content{mcp.Text("appended to " + ref.URN())}, nil
}

func (t *AppendNote) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:           "Append Note",
		ReadOnlyHint:    mcp.HintFalse(),
		DestructiveHint: mcp.HintFalse(),
		OpenWorldHint:   mcp.HintFalse(),
	}
}
