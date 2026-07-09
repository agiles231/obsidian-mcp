package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiles231/obsidian-mcp/internal/vault"
)

func dailyNoteName() string {
	return time.Now().Format("2006-01-02") + ".md"
}

func createDailyNote(dir, content string) string {
	os.MkdirAll(dir, 0755)
	dailyNoteName := dailyNoteName()
	notePath := filepath.Join(dir, dailyNoteName)
	os.WriteFile(notePath, []byte(content), 0644)
	return notePath
}

func setupDailyVault(t *testing.T, dailyConfig string) (*vault.Registry, string) {
	t.Helper()
	dir := t.TempDir()

	// Create .obsidian dir with config
	obsDir := filepath.Join(dir, ".obsidian")
	os.Mkdir(obsDir, 0755)
	if dailyConfig != "" {
		os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(dailyConfig), 0644)
	}

	v, err := vault.Open(vault.Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	reg := vault.NewRegistry()
	reg.Register(v, true)
	return reg, dir
}

func TestDailyNote_ReadExisting(t *testing.T) {
	reg, dir := setupDailyVault(t, `{"folder": "Daily", "format": "YYYY-MM-DD"}`)

	// Create today's note
	dailyDir := filepath.Join(dir, "Daily")
	createDailyNote(dailyDir, "today's content")

	tool := NewDailyNote(reg)
	args, _ := json.Marshal(map[string]any{
		"vault": "test",
		"mode":  "read",
	})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result) == 0 || result[0].Text != "today's content" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDailyNote_Append(t *testing.T) {
	reg, dir := setupDailyVault(t, `{"folder": "Daily", "format": "YYYY-MM-DD"}`)

	// Create today's note
	dailyDir := filepath.Join(dir, "Daily")
	notePath := createDailyNote(dailyDir, "line1")

	tool := NewDailyNote(reg)
	args, _ := json.Marshal(map[string]any{
		"vault":   "test",
		"mode":    "append",
		"content": "line2",
	})

	_, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got, _ := os.ReadFile(notePath)
	if string(got) != "line1\nline2" {
		t.Errorf("got %q", got)
	}
}

func TestDailyNote_DefaultConfig(t *testing.T) {
	reg, dir := setupDailyVault(t, "")

	tool := NewDailyNote(reg)
	args, _ := json.Marshal(map[string]any{
		"vault":   "test",
		"mode":    "append",
		"content": "test",
	})

	_, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	noteName := dailyNoteName()
	notePath := filepath.Join(dir, noteName)
	got, _ := os.ReadFile(notePath)
	if string(got) != "\ntest" {
		t.Errorf("got %q", got)
	}
}

func TestDailyNote_AppendMissingContent(t *testing.T) {
	reg, _ := setupDailyVault(t, `{"folder": "Daily"}`)

	tool := NewDailyNote(reg)
	args, _ := json.Marshal(map[string]any{
		"vault": "test",
		"mode":  "append",
	})

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing content")
	}
}
