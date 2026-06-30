# 1. Direct filesystem vault access

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

The server must read (and later write) an Obsidian vault. Three ways to reach
the vault data were considered:

1. **Obsidian Local REST API plugin** — rich, link-aware, but third-party
   plugin code running inside Obsidian.
2. **Obsidian CLI proxy** — would give Obsidian's link-resolution and
   frontmatter index "for free."
3. **Direct filesystem access** — read the vault's files ourselves.

The project's prime directive is data privacy: vault data must never leave the
machine without explicit permission. The user will not vet third-party plugin
code for exfiltration, which rules out (1). The Obsidian CLI (2) was
investigated: it does have a non-interactive mode (it is not TUI-only, which
was the initial worry), but its capability surface is too limited to justify
binding to an external process and its lifecycle — and we would end up building
our own indexing regardless.

## Decision

Use **direct filesystem access** to the vault. No Obsidian plugin, no external
Obsidian process. Combined with stdio-only transport (no network listener),
everything stays local and auditable.

## Consequences

- **+** Full control; no third-party code to vet; works for non-markdown files
  (attachments, images) too; nothing to bind or firewall.
- **+** The filesystem path is intrinsic, durable metadata that cannot desync.
- **−** We build any indexing, link-resolution, and frontmatter parsing
  ourselves.
- Rejected: the REST plugin (unvetted third-party code) and the CLI proxy
  (limited surface + external-process lifecycle coupling).
