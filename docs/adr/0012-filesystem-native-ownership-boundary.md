# 12. Filesystem-native ownership boundary: own on-disk specs, not plugin runtime

- **Status:** Accepted
- **Date:** 2026-07-10

## Context

As the tool surface grows (frontmatter, search filters, tags, link graph,
templates), a recurring question arises: how much of Obsidian are we
reimplementing, and should we instead drive Obsidian itself?

The only ways to get *plugin-identical* behavior (Dataview, Bases, automatic
link rewriting on rename, Templater scripting) are to run code inside a live
Obsidian process:

| Route | Reality |
|-------|---------|
| "Obsidian CLI" (`obsidian://` URI, toggled by an app setting) | Requires the desktop app running; only triggers actions, doesn't expose a headless API. Already used optimistically by `daily_note` (ADR-0010). |
| Local REST API community plugin | Real Obsidian semantics, but requires the app running with the plugin loaded — a live process and a moving target. |
| Direct filesystem (ADR-0001) | Headless, deterministic, testable. Must re-derive anything Obsidian computes at runtime. |

This server's deployment is agent-first and headless: the vault lives on a
share on a server, and Obsidian — if running at all — runs on a *different*
machine that mounts the share. "Require a live app" therefore defeats the
premise for the common case.

The tension is real: a filesystem-native tool can never be byte-identical to
plugin behavior. We need a principle for *which* Obsidian behaviors are worth
owning and which are out of scope.

## Decision

**Own what has a well-defined, stable on-disk representation. Do not replicate
plugin *runtime* behavior.**

The dividing line is spec stability on disk:

**In scope (ours to own):**
- **Frontmatter** — YAML between leading `---` fences. Fixed location, stable format.
- **Tags** — frontmatter `tags:` lists and inline `#tag`. Well-defined lexical forms.
- **Full-text / phrase / fuzzy search** — operates on file bytes we already read.
- **Frontmatter/tag search filters** — pure metadata predicates once frontmatter is parsed.
- **File/folder layout** — moves, listing, trashing.
- **Core-template instantiation** — static templates with `{{title}}`, `{{date}}`,
  `{{time}}` substitution are spec-stable and headless-friendly.

**Out of scope (stays Obsidian's / plugins'):**
- **Dataview, Bases, rendering, live preview** — runtime query/paint engines.
- **Templater scripting** (`<% tp.* %>`) — a full scripting language; can't run headlessly.
- **Byte-identical link resolution and relevance ranking** — heuristic,
  config-dependent, and a moving target. We aim for *good and predictable for an
  agent*, not identical to the UI.

**Link operations are advisory, not authoritative.** Because we mutate files on
disk (Obsidian is not driving), Obsidian's automatic link updates do not fire.
Rather than fully replicate Obsidian's resolution rules (name vs. path, shortest
-path setting, duplicate names, aliases), the link graph resolves the
unambiguous cases and *reports* the rest. `move`/`rename` rewrite what they can
prove safe and surface what they cannot. Non-corrupting beats clever.

**Optional REST backend as a future escape hatch.** For the few operations where
plugin-identical fidelity genuinely matters (rename with link updates, template
application), allow an optional configured endpoint to a *running* Obsidian
(Local REST API / URI). When present, delegate; otherwise use the filesystem
implementation. This keeps headless-by-default while allowing perfect fidelity
when a live app is available — the same layered pattern ADR-0010 established for
daily notes, generalized.

## Consequences

- **+** Clear, principled answer to "should we build X ourselves": yes iff X has
  a stable on-disk spec.
- **+** Headless, deterministic, testable by default; no dependency on a live app.
- **+** Honest boundaries — we never pretend to be byte-identical to plugins.
- **+** Core templates become usable headlessly (contra the pessimism in ADR-0010,
  which conflated core Templates with Templater).
- **−** Some Obsidian conveniences (Dataview, Templater, guaranteed link updates)
  are unavailable unless the optional REST backend is configured.
- **−** Link rewriting on move/rename is best-effort with reporting, not guaranteed.

## Related

- ADR-0001 (direct filesystem access), ADR-0010 (daily-note CLI/fallback layering),
  ADR-0011 (stdlib-only), ADR-0013 (tool-design principle).
