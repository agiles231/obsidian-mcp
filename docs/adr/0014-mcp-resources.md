# 14. MCP resources for vault ambient context

- **Status:** Accepted
- **Date:** 2026-07-17

## Context

MCP resources provide application-driven, browsable context separate from tools.
Clients can list and attach vault content without the model inventing
`read_file` calls. TODO and the URN spec reserved this surface; tools alone do
not advertise ambient browsability.

Constraints:

- Same privacy and path controls as tools (allow/deny, deny ⇒ not-found).
- URIs must be the canonical `urn:obsidian:` form (ADR-0002 / urn-spec).
- Vaults can be large — full listing must paginate.
- `mcp-stdio-go` previously only advertised tools.

## Decision

1. **Framework** (`mcp-stdio-go`): add `Resources` + optional `ResourceTemplater`,
   `RegisterResources`, and handlers for `resources/list`, `resources/read`,
   `resources/templates/list`. Advertise `capabilities.resources: {}` (no
   `subscribe` / `listChanged` in v1).

2. **Server** (`obsidian-mcp`): implement `internal/resources.VaultResources`:
   - **List** — recursive notes, canvases, attachments (not folders); cursor =
     decimal offset; page size 100.
   - **Read** — parse URN or bare path via `urn.ParseRef`; text for notes/canvases
     and UTF-8 text attachments; base64 blob for binary.
   - **Templates** — `urn:obsidian::{vault}:{type}:{+path}` for direct access.

3. Resource access goes through `Vault` only (ADR-0003); errors collapse to
   `ErrResourceNotFound` (ADR-0005).

## Consequences

- **+** Clients can browse and pin vault context without tool schemas.
- **+** Reuses existing identity and security boundary.
- **−** List materializes a full recursive listing before paging (fine for
  typical vaults; may need index-backed listing later).
- **−** No live update notifications yet.
- **−** Requires a `mcp-stdio-go` release that includes the Resources API.

## Alternatives considered

- **Static resource registration only** — unusable for large vaults.
- **`file://` URIs** — breaks URN-as-identity and multi-vault addressing.
- **List only top-level** — weak ambient discovery; recursive + pagination is
  enough for v1.
