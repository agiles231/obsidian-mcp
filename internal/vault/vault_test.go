package vault

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanVaultRel(t *testing.T) {
	tests := []struct {
		name    string
		rel     string
		want    string
		wantErr error
	}{
		{"simple", "notes/foo.md", "notes/foo.md", nil},
		{"trailing slash", "notes/", "notes", nil},
		{"double slash", "notes//foo.md", "notes/foo.md", nil},
		{"back slash", "notes\\foo.md", "notes/foo.md", nil},
		{"dot segments cleaned", "notes/../other/foo.md", "other/foo.md", nil},
		{"empty", "", "", errInvalid},
		{"dot only", ".", "", errInvalid},
		{"absolute unix", "/etc/passwd", "", errOutsideVault},
		{"absolute windows", "C:\\etc\\passwd", "", errOutsideVault},
		{"escape parent", "../secret", "", errOutsideVault},
		{"escape nested", "a/../../secret", "", errOutsideVault},
		{"dotdot only", "..", "", errOutsideVault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanVaultRel(tt.rel)
			if err != tt.wantErr {
				t.Errorf("cleanVaultRel(%q) error = %v, want %v", tt.rel, err, tt.want)
				return
			}
			if got != tt.want {
				t.Errorf("cleanVaultRel(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestWriteFile_NewFile(t *testing.T) {
	root := t.TempDir()
	v, err := Open(Config{Name: "test", Root: root, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	// The file does not exist yet; WriteFile must be able to create it,
	// not fail with errNotFound because EvalSymlinks can't resolve a
	// not-yet-existing path.
	if err := v.WriteFile(context.Background(), "newfile.md", []byte("hello")); err != nil {
		t.Fatalf("WriteFile on new file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "newfile.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
}

func TestAppendFile_NewFile(t *testing.T) {
	root := t.TempDir()
	v, err := Open(Config{Name: "test", Root: root, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	// append_note's contract is "creates the note if it doesn't exist" -
	// this must not fail with errNotFound on a brand new file.
	if err := v.AppendFile(context.Background(), "newnote.md", []byte("hello")); err != nil {
		t.Fatalf("AppendFile on new file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "newnote.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
}

func TestUnderRoot(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		path   string
		wantOk bool
	}{
		{"inside", "/vault", "/vault/notes/food.md", true},
		{"root itself", "/vault", "/vault", true},
		{"outside parent", "/vault", "/etc/passwd", false},
		{"sibling prefix attack", "/vault", "/vault-evil/foo", false},
		{"escape dotdot", "/vault", "/vault/../etc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := filepath.Clean(tt.root)
			p := filepath.Clean(tt.path)
			_, ok := underRoot(root, p)
			if ok != tt.wantOk {
				t.Errorf("underRoot(%q, %q) ok = %v, want %v", root, p, ok, tt.wantOk)
			}
		})
	}
}
