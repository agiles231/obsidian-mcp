# TODO

## Recently Completed

- [x] Document use cases (`docs/use-cases.md`)
- [x] `NoteRef` → `Ref` refactor (supports all object types)
- [x] `list_objects` tool implementation

## Tests

- [x] Tests for `vault.Registry`
- [x] Tests for `tools.ReadNote`
- [ ] Integration tests with real vault filesystem

## Tools (priority order)

- [x] `read_note` — read note content
- [x] `list_objects` — unified listing with type filters, recursive option (ADR-0007, ADR-0008)
- [ ] `write_note` — create/overwrite a note (ADR-0009)
- [ ] `append_note` — append content to a note (ADR-0009)
- [ ] `search_notes` — full-text search across vault

## Features

- [ ] Indirect config file support (`~/.config/obsidian-mcp/config.toml`)
- [ ] Date-aware path resolution for daily notes
- [ ] Error wrapping with `vault.AgentMessage()` for user-friendly errors

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
