# 4. Build containment on `os.Root` (Go 1.24+)

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

`Vault.resolve` must prevent two escapes: path traversal (`../../etc/passwd`,
absolute paths) and symlinks whose targets leave the vault. The classic
hand-rolled approach — `filepath.Clean` + a string prefix containment check +
`filepath.EvalSymlinks` re-validation — is exactly where subtle security bugs
live: the `/vault` vs `/vault-evil` prefix bug, and a time-of-check-to-time-of-
use (TOCTOU) window between the symlink check and the file open.

Go 1.24 added `os.Root` / `os.OpenRoot`: an openat-based directory handle that
confines every operation to within the root, refusing traversal and escaping
symlinks **at the kernel level, during the actual open** (TOCTOU-resistant).
The current toolchain was 1.23.2.

## Decision

**Target Go 1.24+ and build the vault boundary on `os.Root`.** The deny/allow
policy (ADR-0003) is layered on top.

One subtlety: `os.Root` confines to the vault *root*, but deny-listed
directories live *inside* that root, so a symlink like
`Projects/sneaky.md → private/secret.md` stays within the root and would be
followed — bypassing the deny-list. Therefore the deny check also runs against
the **symlink-resolved real path** (made vault-relative), extending the
"resolve + re-validate containment" rule to "resolve + re-validate containment
**and deny-list**."

The narrow TOCTOU window between resolving the real path (for the deny re-check)
and the `os.Root` open is accepted as **out of scope for v1**: this is a local,
single-user vault, and the realistic threat is prompt-injection steering the
agent — not a hostile process racing symlinks on the user's own machine.
`os.Root` still enforces the vault boundary authoritatively at open time.

## Consequences

- **+** Kernel-enforced, TOCTOU-resistant containment; far less hand-maintained
  security-critical code.
- **−** Requires a toolchain bump to Go 1.24+ (trivial) and `go 1.24` in
  `go.mod`.
- **−** The symlink-into-deny case requires an extra real-path deny check, which
  carries the documented, accepted TOCTOU window.
