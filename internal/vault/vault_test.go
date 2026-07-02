package vault

import (
	"path/filepath"
	"testing"
)

func TestCleanVaultRel(t *testing.T) {
	tests := []struct {
		name string
		rel string
		want string
		wantErr error
	}{
		{"simple", "notes/foo.md", "notes/foo.md", nil},
		{"trailing slash", "notes/", "notes", nil},
		{"double slash", "notes//foo.md", "notes/foo.md", nil},
		{"dot segments cleaned", "notes/../other/foo.md", "other/foo.md", nil},
		{"empty", "", "", errInvalid},
		{"dot only", ".", "", errInvalid},
		{"absolute unix", "/etc/passwd", "", errOutsideVault},
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
				t.Errorf("cleanVaultRel(%q) = %q, want %q", tt.rel, got , tt.want)
			}
		})
	}
}

func TestUnderRoot(t *testing.T) {
	tests := []struct {
		name string
		root string
		path string
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
