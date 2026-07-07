# 7. Unified `list_objects` tool for vault discovery

- **Status:** Accepted
- **Date:** 2026-07-06

## Context

The initial `list_notes` tool only returned markdown files. This left gaps:

1. **No directory discovery.** The LLM couldn't learn folder structure without
   already knowing folder names.
2. **No attachment visibility.** Images, PDFs, and other files were invisible.
3. **Future types.** Obsidian canvases (`.canvas`) and other file types would
   each need their own tool.

A proliferation of tools (`list_notes`, `list_folders`, `list_attachments`,
`list_canvases`) increases cognitive load for the LLM and fragments the API.

## Decision

Replace `list_notes` with a unified **`list_objects`** tool that can list all
vault object types with filtering.

### Object types

| Type | Matches | URN type field |
|------|---------|----------------|
| `note` | `.md` files | `note` |
| `folder` | directories | `folder` |
| `attachment` | non-markdown files (images, PDFs, etc.) | `attachment` |
| `canvas` | `.canvas` files | `canvas` |

### Input schema

```json
{
  "path": "string",       // folder to list, empty = root
  "types": ["string"],    // filter: ["note", "folder", ...], empty = all
  "recursive": "boolean"  // include subdirectories, default false
}
```

### Output structure

Returns a JSON array of objects:

```json
[
  {"urn": "urn:obsidian::vault:note:readme.md", "type": "note", "name": "readme.md"},
  {"urn": "urn:obsidian::vault:folder:Projects", "type": "folder", "name": "Projects"}
]
```

Additional metadata (size, modified time) may be added later as optional fields.

### No separate `list_notes` tool

Rather than maintaining `list_notes` as an alias, we document common filter
patterns. Fewer tools = simpler for the LLM to reason about. The LLM can call
`list_objects` with `types: ["note"]` to get only notes.

## Consequences

- **+** Single tool for all vault discovery — simpler mental model.
- **+** Extensible to new types without API changes.
- **+** LLM can discover folder structure in one call.
- **+** Fewer tools reduces LLM decision complexity.
- **−** Slightly more verbose for the common "just list notes" case.
- **−** URN format needs to accommodate non-note types (folder, attachment) —
  currently `type` in URN is always `note`. May need to extend or use a
  different identifier format for non-note objects.

## Open questions

- Should folders have URNs, or just paths? A folder isn't a "resource" in the
  same way a note is. Could use a simpler `{"type": "folder", "path": "..."}`.
- How to handle the URN `type` field for attachments/canvases — extend the URN
  spec or treat URNs as note-only?
