# 2. Path-based note identity via a `urn:obsidian` URN

- **Status:** Accepted
- **Date:** 2026-06-29
- **Full format spec:** [`docs/urn-spec.md`](../urn-spec.md)

## Context

Notes and sections need a single identifier vocabulary shared across tool
arguments, tool outputs, and (future) MCP resources. Candidate identity bases:

- **Path** — intrinsic, no plugin, legible to an LLM, but breaks on rename/move.
- **Frontmatter stable id** — survives rename, but requires a plugin to mint
  ids (reintroducing the third-party-code dependency ADR-0001 rejected), an
  id→path index with cache invalidation, has no value for non-markdown files,
  and is opaque to the LLM.
- **Wikilink name** (`[[Note]]`) — Obsidian's native grammar, but resolves by
  basename with duplicate-name ambiguity.

Identity is a *naming* concern, not a *locating* one. `obsidian://` is a live,
OS-registered *locator* scheme (it opens the app), so reusing it for identity
would be semantic squatting and would misfire if a string leaked into a
clickable context.

Key reframing: the agent on the other end of stdio is a cloud LLM, so anything
the agent sees has effectively left the machine. Path's rename-breakage is
transient and self-healing for a re-querying agent; the id approach's costs are
permanent.

## Decision

Use a **`urn:` URN over path-based identity**:

```
urn:obsidian:<user>:<vault>:<type>:<identifier>#<anchor>
```

In v1: `user` is empty (reserved, literal `::`), `vault` is populated, `type`
is `note`, `identifier` is a vault-relative path, and the anchor reuses
Obsidian's grammar (`#Heading`, `#Heading#Sub`, `#^blockid`). The resolver is
**liberal in** (accepts a bare vault-relative path or a full URN) and
**canonical out** (always emits the URN). Wikilink/name input is not accepted.

## Consequences

- **+** One identifier vocabulary; no plugin; legible; URN clearly signals
  "name, not a clickable link."
- **+** Additive seams reserved without grammar changes: a future location-
  independent `type=id`, a populated `user`, multi-vault, non-note `type`s.
- **−** No rename stability in v1 (mitigatable later by a `type=id` index and/or
  a search-by-basename fallback on resolve-miss).
- **−** Uses an unregistered `urn:` namespace (internal convention).
- A real `obsidian://open?...` URL may still be emitted as a separate "open in
  Obsidian" convenience — a locator, kept distinct from identity.
