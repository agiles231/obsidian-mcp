package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/agiles231/obsidian-mcp/internal/search"
)

type Config struct {
	Name       string
	Root       string
	ReadAllow  []string
	WriteAllow []string
	Deny       []string
	Logger     *slog.Logger // optional, defaults to slog.Default()
}

type Vault struct {
	name       string
	rootPath   string
	root       *os.Root
	readAllow  patternSet
	writeAllow patternSet
	deny       patternSet
	index      *search.Index
	log        *slog.Logger
}

type ObjectEntry struct {
	Type string // urn.TypeNote, urn.TypeFolder, etc.
	Path string // vault-rel path
	Name string // basename
}
type ListOptions struct {
	Types     map[string]bool // filter by type; nil => all
	Recursive bool
}

type accessKind int

const (
	accessRead accessKind = iota
	accessWrite
)

const maxNoteBytes = 10 << 20 // 10 MiB cap for single note

func Open(cfg Config) (*Vault, error) {
	if cfg.Name == "" {
		return nil, errors.New("vault: empty name")
	}
	// Resolve the root's own symlinks so real-path checks in resolve() align
	rootPath, err := filepath.EvalSymlinks(filepath.Clean(cfg.Root))
	if err != nil {
		return nil, fmt.Errorf("vault: error opening root %s: %v", cfg.Root, err)
	}
	if fi, err := os.Stat(rootPath); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("vault: root not a directory: %s", cfg.Root)
	}
	readAllow, err := compile(cfg.ReadAllow, false)
	if err != nil {
		return nil, err
	}
	writeAllow, err := compile(cfg.WriteAllow, false)
	if err != nil {
		return nil, err
	}
	deny, err := compile(cfg.Deny, true)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}
	return &Vault{
		name: cfg.Name, rootPath: rootPath, root: root, readAllow: readAllow, writeAllow: writeAllow, deny: deny, log: log,
	}, nil
}

func (v *Vault) Name() string { return v.name }

func (v *Vault) BuildSearchIndex(ctx context.Context) error {
	if v.index != nil {
		return nil
	}
	index := search.NewIndex()
	entries, err := v.ListObjects(ctx, "", ListOptions{
		Types:     map[string]bool{"note": true},
		Recursive: true,
	})
	if err != nil {
		return err
	}
	for _, e := range entries {
		bytes, err := v.ReadFile(ctx, e.Path)
		if err != nil {
			return err
		}
		index.Add(e.Path, bytes)
	}
	v.index = index
	return nil
}

func (v *Vault) Search(ctx context.Context, query string, limit int) ([]search.Result, error) {
	if v.index == nil {
		return nil, errors.New("search index not built")
	}
	results, err := v.index.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (v *Vault) ListObjects(ctx context.Context, dir string, opts ListOptions) ([]ObjectEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	startDir := "."
	if dir != "" {
		clean, err := v.resolve(dir, accessRead)
		if err != nil {
			return nil, err
		}
		startDir = clean
	}

	results, err := v.listDir(ctx, startDir, opts)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (v *Vault) ReadFile(ctx context.Context, rel string) (buf []byte, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clean, err := v.resolve(rel, accessRead)
	if err != nil {
		return nil, err
	}
	f, err := v.root.Open(clean)
	if err != nil {
		v.log.Warn("open failed", "path", clean, "err", err)
		return nil, mapFSError(err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			buf = nil
			err = mapFSError(closeErr)
		}
	}()
	data, err := readCapped(f, maxNoteBytes) // prevent local DoS from giant file
	if err != nil {
		v.log.Warn("read failed", "path", clean, "err", err)
		return nil, mapFSError(err)
	}
	return data, nil
}

func (v *Vault) WriteFile(ctx context.Context, rel string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	clean, err := v.resolve(rel, accessWrite)
	if err != nil {
		return err
	}

	// Ensure parent direcotry exists
	dir := path.Dir(clean)
	if dir != "." {
		if err := v.mkdirAll(dir); err != nil {
			return err
		}
	}

	f, err := v.root.Create(clean)
	if err != nil {
		v.log.Warn("create failed", "path", clean, "err", err)
		return mapFSError(err)
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		v.log.Warn("write failed", "path", clean, "err", writeErr)
		return mapFSError(writeErr)
	}
	if closeErr != nil {
		return mapFSError(closeErr)
	}
	return nil
}

func (v *Vault) AppendFile(ctx context.Context, rel string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	clean, err := v.resolve(rel, accessWrite)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := path.Dir(clean)
	if dir != "." {
		if err := v.mkdirAll(dir); err != nil {
			return err
		}
	}

	f, err := v.root.OpenFile(clean, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		v.log.Warn("open for append failed", "path", clean, "err", err)
		return mapFSError(err)
	}

	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		v.log.Warn("append failed", "path", clean, "err", writeErr)
		return mapFSError(writeErr)
	}
	if closeErr != nil {
		return mapFSError(closeErr)
	}
	return nil
}

func (v *Vault) Stat(ctx context.Context, rel string) (fs.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clean, err := v.resolve(rel, accessRead)
	if err != nil {
		return nil, err
	}
	fi, err := v.root.Stat(clean) // os.Root.Stat follows symlinks within root
	if err != nil {
		v.log.Warn("stat failed", "path", clean, "err", err)
		return nil, mapFSError(err)
	}
	return fi, nil
}

// ReadDailyNoteConfig reads the daily notes plugin config.
// Returns zero config (not error) if plugin not configured.
func (v *Vault) ReadDailyNoteConfig() (DailyNoteConfig, error) {
	data, err := v.ReadFile(context.Background(), dailyNotesConfigPath)
	if errors.Is(err, errNotFound) {
		return DailyNoteConfig{}, nil
	}
	if err != nil {
		return DailyNoteConfig{}, err
	}
	var cfg DailyNoteConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DailyNoteConfig{}, err
	}
	return cfg, nil
}

func (v *Vault) listDir(ctx context.Context, dir string, opts ListOptions) ([]ObjectEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := v.root.Open(dir)
	if err != nil {
		v.log.Warn("open dir failed", "path", dir, "err", err)
		return nil, mapFSError(err)
	}
	defer func() { _ = f.Close() }()

	entries, err := f.ReadDir(-1)
	if err != nil {
		v.log.Warn("readir failed", "path", dir, "err", err)
		return nil, mapFSError(err)
	}

	var results []ObjectEntry
	for _, e := range entries {
		name := e.Name()
		var entryPath string
		if dir == "." {
			entryPath = name
		} else {
			entryPath = dir + "/" + name
		}

		// Skip denied entries
		if v.deny.match(entryPath) {
			continue
		}

		// Check allow-list
		if len(v.readAllow) > 0 && !v.readAllow.match(entryPath) {
			continue
		}

		objType := classifyEntry(e, name)

		// Filter by type if specified
		if opts.Types != nil && !opts.Types[objType] {
			// Not including folders, but we are recursing...so recurse
			// to potentially find other matches
			if e.IsDir() && opts.Recursive {
				sub, err := v.listDir(ctx, entryPath, opts)
				if err != nil {
					return nil, err
				}
				results = append(results, sub...)
			}
			continue
		}

		results = append(results, ObjectEntry{
			Type: objType,
			Path: entryPath,
			Name: name,
		})

		// Recurse into folders
		if e.IsDir() && opts.Recursive {
			sub, err := v.listDir(ctx, entryPath, opts)
			if err != nil {
				return nil, err
			}
			results = append(results, sub...)
		}
	}
	return results, nil
}

func (v *Vault) resolve(rel string, op accessKind) (string, error) {
	// gate 1: invalid / absolute / escape
	clean, err := cleanVaultRel(rel)
	if err != nil {
		return "", err
	}
	// gate 2: DENY (string) - deny wins; opaque: looks like a miss
	if v.deny.match(clean) {
		v.log.Warn("access denied by deny-list", "path", clean)
		return "", errNotFound
	}
	// gate 3: ALLOW
	if !v.allowed(clean, op) {
		return "", errNotPermitted
	}

	// gate 4: symlink real-path re-check (containment AND deny) - see ADR-0004
	//
	// The target itself may not exist yet (WriteFile/AppendFile create it),
	// so resolve symlinks against the deepest existing ancestor and rejoin
	// the not-yet-existing remainder rather than failing outright.
	existing, remainder, err := nearestExistingAncestor(filepath.Join(v.rootPath, clean))
	if err != nil {
		return "", errOutsideVault
	}
	real, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", errOutsideVault
	}
	if remainder != "" {
		real = filepath.Join(real, remainder)
	}
	realRel, ok := underRoot(v.rootPath, real)
	if !ok {
		return "", errOutsideVault
	}
	// Check if symlink goes into DENY
	if v.deny.match(realRel) {
		v.log.Warn("access denied by deny-list", "path", realRel)
		return "", errNotFound
	}
	return clean, nil
}

// nearestExistingAncestor walks up from p until it finds a path that
// exists on disk, returning that existing path and the (still-slash-joined,
// not-yet-existing) remainder beneath it. p itself may not exist, but the
// root filesystem always does, so this terminates.
func nearestExistingAncestor(p string) (existing string, remainder string, err error) {
	cur := filepath.Clean(p)
	var parts []string
	for {
		if _, statErr := os.Lstat(cur); statErr == nil {
			return cur, filepath.Join(parts...), nil
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return "", "", statErr
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", "", fs.ErrNotExist
		}
		parts = append([]string{filepath.Base(cur)}, parts...)
		cur = parent
	}
}

func (v *Vault) mkdirAll(rel string) error {
	// Build list of directory paths from root to rel
	parts := strings.Split(rel, "/")
	paths := make([]string, len(parts))
	for i, p := range parts {
		if i == 0 {
			paths[i] = p
		} else {
			paths[i] = paths[i-1] + "/" + p
		}
	}
	// Validate all paths against deny list before mutating state
	if slices.ContainsFunc(paths, v.deny.match) {
		return errNotPermitted
	}

	// Validated; Now create directories
	for _, p := range paths {
		if err := v.root.Mkdir(p, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
			return mapFSError(err)
		}
	}
	return nil
}

func (v *Vault) allowed(clean string, op accessKind) bool {
	switch op {
	case accessRead:
		return len(v.readAllow) == 0 || v.readAllow.match(clean) // empty => all
	case accessWrite:
		return v.writeAllow.match(clean) // empty => none
	}
	return false
}

func classifyEntry(e os.DirEntry, name string) string {
	switch {
	case e.IsDir():
		return "folder"
	case strings.HasSuffix(name, ".md"):
		return "note"
	case strings.HasSuffix(name, ".canvas"):
		return "canvas"
	default:
		return "attachment"
	}
}

// readCapped refuses non-regular files and bounds the read size
func readCapped(f *os.File, max int64) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, errNotFound
	}
	// Read up to max+x so growth path the cap is detected, not truncated
	data, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, errTooLarge
	}
	return data, nil
}

// cleanVaultRel: reject absolute & escaping paths EXPLICITLY (don't silently clamp)
func cleanVaultRel(rel string) (string, error) {
	if rel == "" {
		return "", errInvalid
	}
	s := strings.ReplaceAll(rel, "\\", "/")
	if len(s) >= 2 && s[1] == ':' && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z')) {
		return "", errOutsideVault
	}
	s = filepath.ToSlash(s)
	c := path.Clean(s)
	if path.IsAbs(c) {
		return "", errOutsideVault
	}
	if c == ".." || strings.HasPrefix(c, "../") {
		return "", errOutsideVault
	}
	if c == "." {
		return "", errInvalid
	}
	return c, nil
}

// underRoot: real-path containment, separator aware (no /vault vs /vault-evil bug)
func underRoot(root, p string) (rel string, ok bool) {
	r, err := filepath.Rel(root, p)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) || filepath.IsAbs(r) {
		return "", false
	}
	return filepath.ToSlash(r), true
}
