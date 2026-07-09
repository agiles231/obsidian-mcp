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
- [ ] `tools.DailyNote` — has composition logic, worth testing

### Integration
- [ ] Integration tests with real vault filesystem

## Tools (priority order)

- [x] `read_file` — read note content
- [x] `list_objects` — unified listing with type filters, recursive option (ADR-0007, ADR-0008)
- [x] `write_file` — create/overwrite any file: notes, canvas, etc. (ADR-0009)
- [x] `append_note` — append content to a note (ADR-0009)
- [x] `daily_note` — read/append to today's daily note (ADR-0010)
- [ ] `search_notes` — full-text search across vault

## Features

- [ ] Indirect config file support (`~/.config/obsidian-mcp/config.toml`)
- [ ] Error wrapping with `vault.AgentMessage()` for user-friendly errors

## Skills (slash commands)

### Planned
- [ ] `/save-session` — save conversation summary to vault with templates
- [ ] `/daily` — quick daily note operations (read/append today's note)
- [ ] `/search` — search vault and incorporate context into conversation

### Potential (to assess)
- [ ] `/capture` — quick-capture thought to inbox note
- [ ] `/context` — load relevant notes as conversation context
- [ ] `/weekly-review` — summarize week's daily notes into weekly note
- [ ] `/moc` — generate/update Map of Content linking related notes
- [ ] `/extract` — extract patterns across notes (TODOs, questions)
- [ ] `/backlinks` — find notes that link to a given note
- [ ] `/template` — create note from template with variable substitution
- [ ] `/refactor` — split, merge, restructure notes
- [ ] `/orphans` — find unlinked notes
- [ ] `/frontmatter` — bulk edit/standardize YAML frontmatter
- [ ] `/canvas` — create canvas from conversation or notes

## Future / Deferred (ADR-0009)

- [ ] `write_section` — modify a specific section by anchor
- [ ] `read_section` — read content under a heading
- [ ] `get_sections` — list heading structure of a note
- [ ] Wikilink resolution (`[[Note]]` input)
- [ ] Frontmatter parsing and filtering
- [ ] Backlink discovery ("what links to this note?")
- [ ] MCP resources (browsable context, ambient access)

## Documentation

- [ ] README with usage examples

## Tech Debt

- [ ] Tag `mcp-stdio-go` as `v0.1.0` and pin in `go.mod`
