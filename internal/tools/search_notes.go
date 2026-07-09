package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcp "github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/search"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type SearchNotes struct {
	registry *vault.Registry
}

func NewSearchNotes(r *vault.Registry) *SearchNotes {
	return &SearchNotes{registry: r}
}

func (s *SearchNotes) Name() string { return "search_notes" }
func (s *SearchNotes) Description() string {
	return "Full-text search across vault notes. Returns ranked results with search context."
}

func (s *SearchNotes) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"vault": {
				Type:        "string",
				Description: "Vault name",
			},
			"query": {
				Type:        "string",
				Description: "Search query (space-separated terms, AND semantics)",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum results to return (default 10)",
			},
		},
		Required: []string{"vault", "query"},
	}
}

type searchNotesArgs struct {
	Vault string `json:"vault"`
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (s *SearchNotes) Execute(ctx context.Context, raw json.RawMessage) ([]mcp.Content, error) {
	var args searchNotesArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}

	if args.Limit <= 0 {
		args.Limit = 10
	}

	v, err := s.registry.Get(args.Vault)
	if err != nil {
		return nil, err
	}

	// Build index if not already built
	if err := v.BuildSearchIndex(ctx); err != nil {
		return nil, fmt.Errorf("failed to build search index: %w", err)
	}

	results, err := v.Search(ctx, args.Query, args.Limit)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []mcp.Content{mcp.Text("No results found.")}, nil
	}

	// Build output with context
	var out strings.Builder
	for i, r := range results {
		// Read file for context
		content, err := v.ReadFile(ctx, r.Path)
		if err != nil {
			// File may have been deleted since indexing; skip
			continue
		}

		searchContext := search.SearchContext(string(content), args.Query, 200)

		fmt.Fprintf(&out, "%d. %s (score: %.2f)\n", i+1, r.Path, r.Score)
		fmt.Fprintf(&out, "   %s\n\n", searchContext)
	}
	return []mcp.Content{mcp.Text(out.String())}, nil
}

func (s *SearchNotes) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:           "Search Notes",
		ReadOnlyHint:    mcp.HintTrue(),
		DestructiveHint: mcp.HintFalse(),
		IdempotentHint:  mcp.HintTrue(),
		OpenWorldHint:   mcp.HintFalse(),
	}
}
