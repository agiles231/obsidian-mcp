package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiles231/obsidian-mcp/internal/vault"
)

func setupTestVault(t *testing.T) (*vault.Registry, string) {
	t.Helper()
	root := t.TempDir()

	// Create test files
	notesDir := filepath.Join(root, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "readme.md"), []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(notesDir, "test.md"), []byte("Test content"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := vault.Open(vault.Config{Name: "test", Root: root})
	if err != nil {
		t.Fatal(err)
	}

	r := vault.NewRegistry()
	r.Register(v, true)

	return r, root
}

func TestReadNote_BarePathSuccess(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	args, _ := json.Marshal(readNoteArgs{Ref: "readme.md"})
	content, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(content) != 1 {
		t.Fatalf("got %d content items, want 1", len(content))
	}
	if content[0].Text != "# Hello" {
		t.Errorf("content = %q, want %q", content[0].Text, "# Hello")
	}
}

func TestReadNote_NestedPath(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	args, _ := json.Marshal(readNoteArgs{Ref: "notes/test.md"})
	content, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if content[0].Text != "Test content" {
		t.Errorf("content = %q, want %q", content[0].Text, "# Hello")
	}
}

func TestReadNote_URNSuccess(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	args, _ := json.Marshal(readNoteArgs{Ref: "urn:obsidian::test:note:readme.md"})
	content, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if content[0].Text != "# Hello" {
		t.Errorf("content = %q, want %q", content[0].Text, "# Hello")
	}
}

func TestReadNote_NotFound(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	args, _ := json.Marshal(readNoteArgs{Ref: "nonexistent.md"})
	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestReadNote_VaultNotFound(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	args, _ := json.Marshal(readNoteArgs{Ref: "urn:obsidian::wrong-vault:note:readme.md"})
	_, err := tool.Execute(context.Background(), args)
	if !errors.Is(err, vault.ErrVaultNotFound) {
		t.Errorf("error = %v, want ErrVaultNotFound", err)
	}
}

func TestReadNote_InvalidJSON(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	_, err := tool.Execute(context.Background(), []byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
func TestReadNote_ContextCancelled(t *testing.T) {
	r, _ := setupTestVault(t)
	tool := NewReadNote(r)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args, _ := json.Marshal(readNoteArgs{Ref: "readme.md"})
	_, err := tool.Execute(ctx, args)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
