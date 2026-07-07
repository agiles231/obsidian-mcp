package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	mcp "github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

type DailyNote struct {
	registry *vault.Registry
}

func NewDailyNote(r *vault.Registry) *DailyNote {
	return &DailyNote{registry: r}
}

func (d *DailyNote) Name() string { return "daily_note" }
func (d *DailyNote) Description() string {
	return "Read or append to today's daily note. Creates via Obsidian CLI if available."
}

func (d *DailyNote) Schema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"vault": {
				Type:        "string",
				Description: "Vault name",
			},
			"mode": {
				Type:        "string",
				Enum:        []string{"read", "append", "create"},
				Description: "read: get content, append: add to end, create: ensure exists (uses Obsidian CLI for templates)",
			},
			"content": {
				Type:        "string",
				Description: "Content to append (required for append mode)",
			},
		},
		Required: []string{"vault", "mode"},
	}
}

type dailyNoteArgs struct {
	Vault   string `json:"vault"`
	Mode    string `json:"mode"`
	Content string `json:"content"`
}

func (d *DailyNote) Execute(ctx context.Context, raw json.RawMessage) ([]mcp.Content, error) {
	var args dailyNoteArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	v, err := d.registry.Get(args.Vault)
	if err != nil {
		return nil, err
	}

	cfg, err := v.ReadDailyNoteConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read daily notes config: %w", err)
	}

	today := time.Now()
	path := v.ResolveDailyNotePath(cfg, today)

	switch args.Mode {
	case "read":
		return d.read(ctx, v, path)
	case "append":
		if args.Content == "" {
			return nil, errors.New("content required for append mode")
		}
		return d.append(ctx, v, path, args.Content)
	case "create":
		return d.create(ctx, v, path)
	default:
		return nil, fmt.Errorf("unknown mode: %s", args.Mode)
	}
}

func (d *DailyNote) Annotations() mcp.Annotations {
	return mcp.Annotations{
		Title:           "Daily Note",
		ReadOnlyHint:    mcp.HintFalse(),
		DestructiveHint: mcp.HintFalse(),
		IdempotentHint:  mcp.HintFalse(),
		OpenWorldHint:   mcp.HintTrue(),
	}
}

func (d *DailyNote) read(ctx context.Context, v *vault.Vault, path string) ([]mcp.Content, error) {
	data, err := v.ReadFile(ctx, path)
	if err != nil {
		return nil, err
	}
	return []mcp.Content{mcp.Text(string(data))}, nil
}

func (d *DailyNote) append(ctx context.Context, v *vault.Vault, path string, content string) ([]mcp.Content, error) {
	text := "\n" + content
	if err := v.AppendFile(ctx, path, []byte(text)); err != nil {
		return nil, err
	}
	return []mcp.Content{mcp.Text(fmt.Sprintf("Appended to %s", path))}, nil
}

func (d *DailyNote) create(ctx context.Context, v *vault.Vault, path string) ([]mcp.Content, error) {
	// Check if daily already exists
	if _, err := v.Stat(ctx, path); err == nil {
		return []mcp.Content{mcp.Text(fmt.Sprintf("Daily note exists: %s", path))}, nil
	}

	// Try Obsidian CLI first
	if err := openObsidianDaily(v.Name()); err == nil {
		// Give Obsidian time to create the file
		time.Sleep(500 * time.Millisecond)
		if _, err := v.Stat(ctx, path); err == nil {
			return []mcp.Content{mcp.Text(fmt.Sprintf("Created via Obsidian: %s", path))}, nil
		}
	}
	if err := v.WriteFile(ctx, path, []byte("")); err != nil {
		return nil, err
	}
	return []mcp.Content{mcp.Text(fmt.Sprintf("Create (no template): %s", path))}, nil
}

func openObsidianDaily(vaultName string) error {
	uri := fmt.Sprintf("obsidian://daily?vault=%s", vaultName)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", uri)
	case "linux":
		cmd = exec.Command("xdg-open", uri)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", uri)
	default:
		return errors.New("unsupported platform")
	}
	return cmd.Run()
}
