# 3. A single `Vault` abstraction is the sole gateway to disk

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

The prime directive requires comprehensive access control — a readable/writable
allow-list plus a sensitive-directory deny-list — enforced *before* any tool
touches disk. If individual tools call `os.ReadFile` directly, the security
boundary is duplicated, easy to bypass, and easy to get subtly wrong per-tool.

## Decision

Introduce **one `Vault` type that owns all filesystem I/O**. Tools never touch
disk directly; they call `Vault.ReadFile` / `Vault.Stat` (and later write
methods). The path-validation routine (`resolve`) is private — a validated
absolute path never escapes into tool code, so the gate cannot be bypassed.

Allow/deny matching uses a **custom glob matcher with `**` support** (zero
third-party dependency, fully auditable, consistent with ADR-0001's posture).
Matching is **fail-closed**:

- **Deny-list:** a bare entry covers its whole subtree (`private` denies
  `private` *and* `private/**`), so protection can't be defeated by nesting.
- **Allow-list:** no auto-expansion — least privilege, stays conservative.
- **Deny wins** over allow.

## Consequences

- **+** A single, auditable security boundary that cannot be bypassed; one place
  for access logging.
- **+** Custom glob keeps the dependency surface at zero and the matching logic
  reviewable (a matching bug here is a security bug).
- **−** All disk access funnels through one type — an acceptable, intended
  constraint.
- See ADR-0004 for the containment mechanism and ADR-0005 for how refusals are
  reported.
