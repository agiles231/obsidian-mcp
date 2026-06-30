package vault

import (
	"errors"
	"io/fs"
)

var (
	errInvalid      = errors.New("invalid indentifier")
	errOutsideVault = errors.New("path is outside the vault")
	errNotPermitted = errors.New("path is not in the permitted set")
	errNotFound     = errors.New("not found")
	errTooLarge     = errors.New("file too large")
)

func AgentMessage(err error) string {
	switch {
	case errors.Is(err, errInvalid):
		return "invalid path or identifier"
	case errors.Is(err, errOutsideVault):
		return "path is outside the vault"
	case errors.Is(err, errNotPermitted):
		return "path is not in the permitted set"
	case errors.Is(err, errNotFound):
		return "not found"
	case errors.Is(err, errTooLarge):
		return "file is too large to read"
	default:
		return "request failed" // generic message; hide internal errors
	}
}

// mapFSError normalizes any filesystem-layer error into a Vault sentinel.
// Vault sentinels pass through unchanged; raw OS errors collapse so that
// "missing", "denied", and "no permission" are indistinguishable (ADR-0005).
func mapFSError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errInvalid), errors.Is(err, errOutsideVault),
		errors.Is(err, errNotPermitted), errors.Is(err, errNotFound),
		errors.Is(err, errTooLarge):
		return err // already a Vault sentinel
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, fs.ErrPermission):
		return errNotFound
	default:
		return errOutsideVault // os.Root containment escape (TOCTOU) / unexpected
	}
}
