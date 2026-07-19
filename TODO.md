# TODO

## Recently Completed

- [x] Document use cases (`docs/use-cases.md`)
- [x] `NoteRef` → `Ref` refactor (supports all object types)
- [x] `list_objects` tool implementation

## Tests

### Existing
- [x] `vault.Registry`
- [x] `vault.Vault` (core operations)
- [x] `vault.patternSet` (glob matching)
- [x] `vault.errors` (error mapping)
- [x] `urn.Parse` / `urn.Ref`
- [x] `tools.ReadFile`

### Missing — vault layer
- [x] `vault.formatMomentDate` — Moment.js → Go time conversion
- [x] `vault.ResolveDailyNotePath` — path construction from config + date
- [x] `vault.ReadDailyNoteConfig` — JSON parsing, missing file handling
- [x] `vault.WriteFile` — create, overwrite, mkdir behavior
- [x] `vault.AppendFile` — append to existing, create new
- [x] `vault.ListObjects` — type filtering, recursion, deny-list

### Missing — tools layer
- [x] `tools.ReadFile` — existing
- [~] `tools.WriteFile` — thin wrapper, covered by vault tests
- [~] `tools.AppendNote` — thin wrapper, covered by vault tests
- [~] `tools.ListObjects` — thin wrapper, covered by vault tests
- [x] `tools.DailyNote` — has composition logic, worth testing

### Integration
- [ ] Integration tests with real vault filesystem

## Tools (priority order)

- [x] `read_file` — read note content
- [x] `list_objects` — unified listing with type filters, recursive option (ADR-0007, ADR-0008)
- [x] `write_file` — create/overwrite any file: notes, canvas, etc. (ADR-0009)
- [x] `append_note` — append content to a note (ADR-0009)
- [x] `daily_note` — read/append to today's daily note (ADR-0010)
- [x] `search_notes` — full-text search across vault

## Features

- [ ] Indirect config file support (`~/.config/obsidian-mcp/config.toml`)
- [ ] Error wrapping with `vault.AgentMessage()` for user-friendly errors

## Skills (slash commands)

See [`skills/`](skills/) — prompt packages for agents using obsidian-mcp.
Wire discovery via `[skills] paths` or a symlink into `.claude/skills` /
`.grok/skills` (see `skills/README.md`).

### Planned
- [x] `/save-session` — save conversation summary to vault with templates
- [x] `/daily` — quick daily note operations (read/append today's note)
- [x] `/search` — search vault and incorporate context into conversation

### Potential (first draft; refine with use)
- [x] `/capture` — quick-capture thought to inbox note
- [x] `/context` — load relevant notes as conversation context
- [x] `/weekly-review` — summarize week's daily notes into weekly note
- [x] `/moc` — generate/update Map of Content linking related notes
- [x] `/extract` — extract patterns across notes (TODOs, questions)
- [x] `/backlinks` — find notes that link to a given note (approx via search)
- [x] `/template` — create note from template with variable substitution
- [x] `/refactor` — split, merge, restructure notes
- [x] `/orphans` — find unlinked notes (approx)
- [x] `/frontmatter` — bulk edit/standardize YAML frontmatter
- [x] `/canvas` — create canvas from conversation or notes

## Planned tools & subsystems (ADR-0012, ADR-0013)

Sequenced by *foundation*, not by tool. Admission rule (ADR-0013): a specific
tool must carry an invariant the primitive can't; otherwise it's sugar — reject.
Ownership rule (ADR-0012): own on-disk specs; link ops are advisory.

### Foundation A — Frontmatter engine (`internal/frontmatter`)
Pure parse ↔ render, body-preserving. Unlocks most of the list below.
- [ ] `internal/frontmatter` package: `Split` / `Patch` / `Render` (+ tests:
      empty file, no-FM, CRLF, list values, quoting, unset-last-key, body byte-preserved)
- [ ] `vault.UpdateFrontmatter(ctx, rel, set, unset)` (not-found ⇒ new block)
- [ ] `patch_note_frontmatter` tool — merge; preserves body + untouched keys **(core)**
- [ ] `read_note_frontmatter` tool — return parsed, structured metadata (no body round-trip)
- [ ] `tag_note` / `remove_tag` — **conditional**: admit only if they canonicalize
      (dedupe, nested tags, frontmatter-list vs inline `#tag`); else fold into patch
- [x] ~~`write_note_frontmatter`~~ — **folded** into `patch_note_frontmatter` (sugar; ADR-0013)
- [x] ~~`prepend_note`~~ — **dropped** (sugar over append; frontmatter need covered by patch; ADR-0013)

### Foundation B — Link graph (advisory; ADR-0012)
Parsing links is easy; *resolution parity* is not — resolve unambiguous cases,
report the rest. Never corrupt.
- [ ] Link parser: wikilink / embed / markdown-link, with `#heading` and `#^block` suffixes
- [ ] Backlink index + resolution (unique-name cases resolved; duplicates/aliases flagged)
- [ ] `get_backlinks` tool
- [ ] `move_note` — rewrite provably-safe references, report the unsafe ones
- [ ] `move_folder` — bulk `move_note`
- [ ] (later) optional REST backend delegation for rename-with-link-updates (ADR-0012)

### Independent
- [ ] `trash_note` — recoverable delete to `.trash` (never hard delete) **(small, high utility)**
- [ ] `list_tags` — enumerate tags + counts (schema hygiene)
- [ ] `create_from_template` — instantiate a core template with `{{title}}`/`{{date}}`/`{{time}}`
      substitution (headless-safe; Templater `<% %>` explicitly out of scope, ADR-0012)
- [ ] `list_objects` field projection — opt-in `fields: [type, status]` returns only chosen
      frontmatter keys (triage, e.g. "notes missing `type`"); avoids dumping full frontmatter

## Search upgrades (ADR-0011, ADR-0012)

Current state: BM25 with positional postings already indexed (`internal/search`).
- [ ] Frontmatter / tag **filters** (needs Foundation A) — highest value, nearly free
- [ ] Phrase matching — positions already stored; just consume them
- [ ] Boolean / field-scoped queries (AND/OR/NOT)
- [ ] Fuzzy matching — bounded edit-distance / trigram; flag-gated so exact stays fast
- [ ] Expose options on `search_notes` (mode: exact|phrase|fuzzy, filters)
- Non-goal: Obsidian-identical ranking / search-operator syntax (ADR-0012)

## Future / Deferred (ADR-0009)

- [ ] Section trio (add together or not at all — ADR-0009): `get_sections` /
      `read_section` / `write_section`
	- [ ] `write_section` — modify a specific section by anchor
	- [ ] `read_section` — read content under a heading
	- [ ] `get_sections` — list heading structure of a note
- [ ] Wikilink resolution as tool input (`[[Note]]` → path) — rides on Foundation B
- [ ] Frontmatter parsing and filtering
- [ ] Backlink discovery ("what links to this note?")
- [x] MCP resources (browsable context, ambient access)

## Documentation

- [x] README with usage examples

## Tech Debt

- [x] Tag `mcp-stdio-go` as `v0.1.0` and pin in `go.mod`
- [ ] Search index never invalidates: `BuildSearchIndex` guards on `v.index != nil`,
      so results go stale after any write. Add incremental update (`Index.Add`/`Remove`
      on write/trash/move) and/or invalidation. (Index is already lazy — built on first
      `search_notes`, not at startup — so laziness itself is not the issue.)
