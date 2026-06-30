package vault

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	log        *slog.Logger
}

func Open(cfg Config) (*Vault, error) {
	if cfg.Name == "" {
		return nil, errors.New("vault: empty name")
	}
	// Resolve the root's own symlinks so real-path checks in resolve() align
	rootPath, err := filepath.EvalSymlinks(filepath.Clean(cfg.Root))
	if err != nil {

	}
	if fi, err := os.Stat(rootPath); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("vault: root not a directory: %s", cfg.Root)
	}
	readAllow, err := compile(cfg.ReadAllow, false)
	if err != nil {
		return nil, err
	}
	writeAllow, err := compile(cfg.ReadAllow, false)
	if err != nil {
		return nil, err
	}
	deny, err := compile(cfg.ReadAllow, false)
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

const maxNoteBytes = 10 << 20 // 10 MiB cap for single note
func (v *Vault) ReadFile(ctx context.Context, rel string) ([]byte, error) {
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
	defer f.Close()
	data, err := readCapped(f, maxNoteBytes) // prevent local DoS from giant file
	if err != nil {
		v.log.Warn("read failed", "path", clean, "err", err)
		return nil, mapFSError(err)
	}
	return data, nil
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

type accessKind int

const (
	accessRead accessKind = iota
	accessWrite
)

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
	real, err := filepath.EvalSymlinks(filepath.Join(v.rootPath, clean))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", errNotFound
		}
		return "", errOutsideVault
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

func (v *Vault) allowed(clean string, op accessKind) bool {
	switch op {
	case accessRead:
		return len(v.readAllow) == 0 || v.readAllow.match(clean) // empty => all
	case accessWrite:
		return v.writeAllow.match(clean) // empty => none
	}
	return false
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
	s := filepath.ToSlash(rel)
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
