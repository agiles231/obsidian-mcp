package resources

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/urn"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

func openTestRegistry(t *testing.T, root string) *vault.Registry {
	t.Helper()
	v, err := vault.Open(vault.Config{
		Name: "test-vault",
		Root: root,
		Deny: []string{".obsidian", "private"},
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	reg := vault.NewRegistry()
	if err := reg.Register(v, true); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return reg
}

func TestVaultResourcesListAndRead(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Hello.md"), "# Hello\n")
	mustWrite(t, filepath.Join(root, "Projects", "Plan.md"), "plan\n")
	mustWrite(t, filepath.Join(root, "board.canvas"), `{"nodes":[]}`)
	mustWrite(t, filepath.Join(root, "private", "secret.md"), "nope\n")
	mustWrite(t, filepath.Join(root, "img.bin"), string([]byte{0x00, 0x01, 0xff}))

	r := NewVaultResources(openTestRegistry(t, root))
	r.pageSize = 2 // exercise pagination

	ctx := context.Background()

	page1, next, err := r.List(ctx, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}
	if next == "" {
		t.Fatal("expected next cursor after first page")
	}

	page2, next2, err := r.List(ctx, next)
	if err != nil {
		t.Fatalf("List page2: %v", err)
	}
	if next2 != "" && len(page2) == 0 {
		t.Fatal("empty page with next cursor")
	}

	// Collect all URIs across pages.
	all := append([]mcp.ResourceDescriptor{}, page1...)
	cursor := next
	for cursor != "" {
		var page []mcp.ResourceDescriptor
		page, cursor, err = r.List(ctx, cursor)
		if err != nil {
			t.Fatalf("List remaining: %v", err)
		}
		all = append(all, page...)
	}

	uris := map[string]mcp.ResourceDescriptor{}
	for _, d := range all {
		uris[d.URI] = d
		// Denied paths must never appear.
		if d.Title == "private/secret.md" || d.Name == "secret.md" {
			t.Errorf("denied path listed as resource: %+v", d)
		}
	}

	// Read a note by URN.
	noteRef := urn.Ref{Vault: "test-vault", Type: urn.TypeNote, Path: "Hello.md"}
	noteURI := noteRef.URN()
	if _, ok := uris[noteURI]; !ok {
		t.Fatalf("Hello.md not listed; got URIs: %v", keys(uris))
	}
	contents, err := r.Read(ctx, noteURI)
	if err != nil {
		t.Fatalf("Read note: %v", err)
	}
	if len(contents) != 1 || contents[0].Text != "# Hello\n" {
		t.Errorf("Read note contents = %+v", contents)
	}
	if contents[0].MimeType != "text/markdown" {
		t.Errorf("mime = %q, want text/markdown", contents[0].MimeType)
	}

	// Bare path also works for read.
	contents, err = r.Read(ctx, "Projects/Plan.md")
	if err != nil {
		t.Fatalf("Read bare path: %v", err)
	}
	if len(contents) != 1 || contents[0].Text != "plan\n" {
		t.Errorf("bare path contents = %+v", contents)
	}

	// Denied path looks like not found.
	_, err = r.Read(ctx, "private/secret.md")
	if err != mcp.ErrResourceNotFound {
		t.Errorf("denied read err = %v, want ErrResourceNotFound", err)
	}

	// Missing path.
	_, err = r.Read(ctx, "nope.md")
	if err != mcp.ErrResourceNotFound {
		t.Errorf("missing read err = %v, want ErrResourceNotFound", err)
	}

	// Binary attachment returns blob.
	binRef := urn.Ref{Vault: "test-vault", Type: urn.TypeAttachment, Path: "img.bin"}
	binURI := binRef.URN()
	contents, err = r.Read(ctx, binURI)
	if err != nil {
		t.Fatalf("Read binary: %v", err)
	}
	if len(contents) != 1 || contents[0].Blob == "" || contents[0].Text != "" {
		t.Errorf("binary contents = %+v, want blob only", contents)
	}

	// Templates.
	tmpls, _, err := r.ListTemplates(ctx, "")
	if err != nil || len(tmpls) != 1 {
		t.Fatalf("ListTemplates: %v %+v", err, tmpls)
	}
	if tmpls[0].URITemplate == "" {
		t.Error("empty uriTemplate")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func keys(m map[string]mcp.ResourceDescriptor) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
