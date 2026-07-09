package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteFile_Create(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	// The file does not exist yet; WriteFile must be able to create it,
	// not fail with errNotFound because EvalSymlinks can't resolve a
	// not-yet-existing path.
	if err := v.WriteFile(context.Background(), "newfile.md", []byte("hello")); err != nil {
		t.Fatalf("WriteFile on new file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "newfile.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
}

func TestWriteFile_Overwrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "exist.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.WriteFile(context.Background(), "exist.md", []byte("new")); err != nil {
		t.Fatalf("WriteFile on existing file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "exist.md"))
	if err != nil {
		t.Fatalf("reading existing file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
}

func TestWriteFile_MkdirAll(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.WriteFile(context.Background(), "a/b/c/note.md", []byte("deep")); err != nil {
		t.Fatalf("WriteFile on existing file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "a/b/c/note.md"))
	if err != nil {
		t.Fatalf("reading nested file: %v", err)
	}
	if string(got) != "deep" {
		t.Errorf("content = %q, want %q", got, "new")
	}
}

func TestWriteFile_DenyList(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}, Deny: []string{"private"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	err = v.WriteFile(context.Background(), "private/secret.md", []byte("nope"))
	if err == nil {
		t.Error("expected error for denied path")
	}
}

func TestWriteFile_NotAllowed(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"notes/**"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	err = v.WriteFile(context.Background(), "other/file.md", []byte("nope"))
	if err == nil {
		t.Error("expected error for path not in WriteAllow")
	}
}

func TestAppendFile_Existing(t *testing.T) {
	dir := t.TempDir()
	// Write file directly
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.AppendFile(context.Background(), "note.md", []byte("line2\n")); err != nil {
		t.Fatalf("AppendFile on existing file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "note.md"))
	if err != nil {
		t.Fatalf("reading appended file: %v", err)
	}
	if string(got) != "line1\nline2\n" {
		t.Errorf("got = %q, want %q", got, "hello")
	}
}

func TestAppendFile_Create(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.AppendFile(context.Background(), "newnote.md", []byte("hello")); err != nil {
		t.Fatalf("AppendFile on new file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "newnote.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("got = %q, want %q", got, "hello")
	}
}

func TestAppendFile_MkdirAll(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.AppendFile(context.Background(), "a/b/note.md", []byte("deep")); err != nil {
		t.Fatalf("AppendFile on deep file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "a/b/note.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}
	if string(got) != "deep" {
		t.Errorf("got = %q, want %q", got, "deep")
	}
}

func TestAppendFile_DenyList(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir, WriteAllow: []string{"**"}, Deny: []string{"private/**"}})
	if err != nil {
		t.Fatal(err)
	}

	err = v.AppendFile(context.Background(), "private/secret.md", []byte("nope"))
	if err == nil {
		t.Fatal("expected error for denied path")
	}
}

func TestListObject_Basic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "note.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "image.md"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(dir, "folder"), 0755)

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	entries, err := v.ListObjects(context.Background(), "", ListOptions{})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("got %d entries, want 3", len(entries))
	}
}

func TestListObject_TypeFilter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "img.png"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	entries, err := v.ListObjects(context.Background(), "", ListOptions{
		Types: map[string]bool{"note": true},
	})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("got %d notes, want 2 notes", len(entries))
	}

	for _, e := range entries {
		if e.Type != "note" {
			t.Errorf("unexpected type %q", e.Type)
		}
	}
}

func TestListObject_Recursive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "top.md"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(dir, "a/b"), 0755)
	os.WriteFile(filepath.Join(dir, "a/mid.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "a/b/deep.md"), []byte("test"), 0644)

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	entries, err := v.ListObjects(context.Background(), "", ListOptions{
		Types:     map[string]bool{"note": true},
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("got %d notes, want 3", len(entries))
	}
}

func TestListObject_DenyList(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "public.md"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(dir, "private"), 0755)
	os.WriteFile(filepath.Join(dir, "private/secret.md"), []byte("test"), 0644)

	v, err := Open(Config{Name: "test", Root: dir, Deny: []string{"private/**"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	entries, err := v.ListObjects(context.Background(), "", ListOptions{
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Path, "private") {
			t.Errorf("denied path leaked: %s", e.Path)
		}
	}
}

func TestListObject_Subdir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "notes"), 0755)
	os.WriteFile(filepath.Join(dir, "notes/a.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "notes/b.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "other.md"), []byte("test"), 0644)

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	entries, err := v.ListObjects(context.Background(), "notes", ListOptions{})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("got %d notes, want 2", len(entries))
	}
}

func TestReadDailyNoteConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	cfg, err := v.ReadDailyNoteConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg != (DailyNoteConfig{}) {
		t.Errorf("expected zero config, got %+v", cfg)
	}
}

func TestReadDailyNoteConfig_Valid(t *testing.T) {
	tests := []struct {
		name         string
		fileContents string
		want         DailyNoteConfig
	}{
		{
			"full config",
			`{"folder": "Daily", "format": "YYYY-MM-DD", "template": "Templates/Daily"}`,
			DailyNoteConfig{
				Folder:   "Daily",
				Format:   "YYYY-MM-DD",
				Template: "Templates/Daily",
			},
		},
		{
			"partial config",
			`{"folder": "Journal"}`,
			DailyNoteConfig{
				Folder:   "Journal",
				Format:   "",
				Template: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			// Create .obsidian/daily-notes.json
			obsDir := filepath.Join(dir, ".obsidian")
			if err := os.Mkdir(obsDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(tt.fileContents), 0644); err != nil {
				t.Fatal(err)
			}

			v, err := Open(Config{Name: "test", Root: dir})
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			cfg, err := v.ReadDailyNoteConfig()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if cfg.Folder != tt.want.Folder || cfg.Format != tt.want.Format || cfg.Template != tt.want.Template {
				t.Errorf("unexpected config: %+v", cfg)
			}
		})
	}
}

func TestReadDailyNoteConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	// Create .obsidian/daily-notes.json
	obsDir := filepath.Join(dir, ".obsidian")
	if err := os.Mkdir(obsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte("{not valid}"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	_, err = v.ReadDailyNoteConfig()
	if err == nil {
		t.Error("expected error for invalid JSON")
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
