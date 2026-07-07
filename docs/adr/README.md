# Architecture Decision Records

These ADRs capture the significant, hard-to-reverse decisions behind
`obsidian-mcp` and *why* they were made, so a future contributor (human or
agent) can understand the reasoning without re-deriving it.

Format: [Michael Nygard's template](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) —
**Context → Decision → Consequences**, plus Status and Date. One decision per
record. When a decision changes, add a new ADR that supersedes the old one
(mark the old one `Superseded by ADR-NNNN`); don't rewrite history.

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-direct-filesystem-vault-access.md) | Direct filesystem vault access | Accepted |
| [0002](0002-urn-path-identity.md) | Path-based note identity via a `urn:obsidian` URN | Accepted |
| [0003](0003-vault-sole-disk-gateway.md) | A single `Vault` abstraction is the sole gateway to disk | Accepted |
| [0004](0004-os-root-containment.md) | Build containment on `os.Root` (Go 1.24+) | Accepted |
| [0005](0005-refusal-error-model.md) | Refusal error model (split taxonomy, deny ⇒ not-found) | Accepted |
| [0006](0006-threat-model-security-boundaries.md) | Threat model and security boundaries | Accepted |
| [0007](0007-list-objects-unified-listing.md) | Unified `list_objects` tool for vault discovery | Accepted |
| [0008](0008-object-type-taxonomy.md) | Object type taxonomy for vault listing | Accepted |
