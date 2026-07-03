package urn

import (
	"fmt"
	"errors"
	"testing"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		name string
		input string
		wantVault string
		wantPath string
		wantAnc Anchor
		wantErr error
	}{
		{
			name: "simple note",
			input: "urn:obsidian::my-vault:note:Projects/foo.md",
			wantVault: "my-vault",
			wantPath: "Projects/foo.md",
		},
		{
			name: "case insensitive prefix",
			input: "URN:OBSIDIAN::my-vault:note:foo.md",
			wantVault: "my-vault",
			wantPath: "foo.md",
		},
		{
			name: "heading anchor",
			input: "urn:obsidian::my-vault:note:note.md#Design",
			wantVault: "my-vault",
			wantPath: "note.md",
			wantAnc: Anchor{Headings: []string{"Design"}},
		},
		{
			name: "nested heading anchor",
			input: "urn:obsidian::my-vault:note:note.md#Design#Identity",
			wantVault: "my-vault",
			wantPath: "note.md",
			wantAnc: Anchor{Headings: []string{"Design", "Identity"}},
		},
		{
			name: "block ref anchor",
			input: "urn:obsidian::my-vault:note:daily.md#^a1b2c3",
			wantVault: "my-vault",
			wantPath: "daily.md",
			wantAnc: Anchor{BlockID: "a1b2c3"},
		},
		{
			name: "percent-encoded path",
			input: "urn:obsidian::my-vault:note:Meeting%20Notes/Q3%20Planning.md",
			wantVault: "my-vault",
			wantPath: "Meeting Notes/Q3 Planning.md",
		},
		{
			name: "percent-encoded vault",
			input: "urn:obsidian::my%20vault:note:foo.md",
			wantVault: "my vault",
			wantPath: "foo.md",
		},
		{
			name: "percent-encoded heading",
			input: "urn:obsidian::vault:note:foo.md#Section%20One",
			wantVault: "vault",
			wantPath: "foo.md",
			wantAnc: Anchor{Headings: []string{"Section One"}},
		},
		{
			name: "missing fields",
			input: "urn:obsidian::vault:note",
			wantErr: ErrInvalidURN,
		},
		{
			name: "non-empty user field",
			input: "urn:obsidian:alice:vault:note:foo.md",
			wantErr: ErrUserReserved,
		},
		{
			name: "empty vault",
			input: "urn:obsidian:::note:foo.md",
			wantErr: ErrEmptyVault,
		},
		{
			name: "unknown type",
			input: "urn:obsidian::vault:canvas:foo.canvas",
			wantErr: ErrUnknownType,
		},
		{
			name: "empty path",
			input: "urn:obsidian::vault:note:",
			wantErr: ErrEmptyPath,
		},
		{
			name: "empty path segment",
			input: "urn:obsidian::vault:note:a//b.md",
			wantErr: ErrInvalidURN,
		},
		{
			name: "empty block ID",
			input: "urn:obsidian::vault:note:foo.md#^",
			wantErr: ErrInvalidURN,
		},
		{
			name: "not a URN",
			input: "http://example.com",
			wantErr: ErrEmptyVault, // falls through to parseBare with no default value
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseRef(tt.input, "")
			if err != nil && tt.wantErr == nil {
				t.Errorf("err = %v, want nil", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("err nil, want %v", tt.wantErr)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseRef(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				if ref != nil {
					t.Errorf("ref = %v, want nil", ref)
				} else {
					return
				}
			}
			fmt.Printf("err %v, ref %v\n", err, ref)
			if ref.Vault != tt.wantVault {
				t.Errorf("Vault = %q, want %q", ref.Vault, tt.wantVault)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", ref.Path, tt.wantPath)
			}
			if ref.Anchor.BlockID != tt.wantAnc.BlockID {
				t.Errorf("Anchor.BlockID = %q, want %q", ref.Anchor.BlockID, tt.wantAnc.BlockID)
			}
			if len(ref.Anchor.Headings) != len(tt.wantAnc.Headings) {
				t.Errorf("Anchor.Headings = %q, want %q", ref.Anchor.Headings, tt.wantAnc.Headings)
			} else {
				for i, h := range ref.Anchor.Headings {
					if h != tt.wantAnc.Headings[i] {
						t.Errorf("Anchor.Headings[%d] = %q, want %q", i, h, tt.wantAnc.Headings[i])
					}
				}
			}
		})
	}
}
