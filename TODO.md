# TODO

## In Progress

- [x] Document use cases (`docs/use-cases.md`)
- [x] Update `CONTEXT.md` to reflect current state

## Tests

- [x] Tests for `vault.Registry`
- [x] Tests for `tools.ReadNote`
- [ ] Integration tests with real vault filesystem

## Tools (priority order)

- [x] `read_note` — read note content
- [x] `list_notes` — list notes in a folder (flat listing) — **superseded by list_objects**
- [ ] `list_objects` — unified listing with type filters, recursive option (ADR-0007)
- [ ] `write_note` — create/overwrite a note
- [ ] `append_note` — append content to a note
- [ ] `search_notes` — full-text search across vault

## Features

- [ ] Anchor handling in `read_note` (extract heading/block sections)
- [ ] Indirect config file support (`~/.config/obsidian-mcp/config.toml`)
- [ ] Date-aware path resolution for daily notes
- [ ] Error wrapping with `vault.AgentMessage()` for user-friendly errors

## Future / Deferred

- [ ] `update_section` — modify a specific section by anchor
- [ ] `get_note_sections` — list heading structure of a note
- [ ] Wikilink resolution (`[[Note]]` input)
- [ ] Frontmatter parsing and filtering
- [ ] Backlink discovery ("what links to this note?")
- [ ] MCP resources (browsable context, ambient access)

## Documentation

- [ ] README with usage examples
- [ ] ADR index in `docs/adr/README.md` — add 0006

## Tech Debt

- [ ] Tag `mcp-stdio-go` as `v0.1.0` and pin in `go.mod`
