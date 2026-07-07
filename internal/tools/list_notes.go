package tools

import (
	"context"
	"encoding/json"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type ListNotes struct {
	registry *vault.Registry
}

func NewListNotes(r *vault.Registry) *ListNotes {
	return &ListNotes{registry: r}
}

func (t *ListNotes) Name() string {
	return "list_notes"
}

func (t *ListNotes) Description() string {
	return "List notes in a folder. Returns URNs for each note. Pass an empty path to list teh vault root."
}

func (t *ListNotes) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"path": {
				Type: "string",
				Description: "Folder path relative to the vault root. Empty string lists the root.",
			},
		},
		Required: []string{},
	}
}

type listNotesArgs struct {
	Path string `json:"path"`
}

func (t *ListNotes) Execute(ctx context.Context, args json.RawMessage) ([]mcp.Content, error) {
	var a listNotesArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}

	v, err := t.registry.Default()
	if err != nil {
		return nil, err
	}

	paths, err := v.ReadDir(ctx, a.Path)
	if err != nil {
		return nil, err
	}

	// Convert paths to URNs
	urns := []string{}
	for _, p := range paths {
		ref := urn.Ref{Vault: v.Name(), Type: urn.TypeNote, Path: p}
		urns = append(urns, ref.URN())
	}

	// Return as JSON array
	result, err := json.Marshal(urns)
	if err != nil {
		return nil, err
	}

	return []mcp.Content{mcp.Text(string(result))}, nil
}

func (t *ListNotes) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title: "List Notes",
		ReadOnlyHint: mcp.HintTrue(),
		OpenWorldHint: mcp.HintFalse(),
	}
}
