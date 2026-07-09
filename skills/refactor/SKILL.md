---
name: refactor
description: >
  Split, merge, or restructure Obsidian notes using full-file read/write.
  Use when the user runs /refactor, asks to "split this note", "merge these
  notes", "move section to its own note", or restructure vault content.
argument-hint: "[split|merge|restructure] <paths and intent>"
---

# Refactor Notes

Restructure markdown notes with the v1 write model: full-file `read_file` +
`write_file` (no section patch tool).

## Prerequisites

- MCP tools: `read_file`, `write_file`, `append_note`, optionally `list_objects`

## Modes

### Split

1. `read_file` the source.
2. Agree on split boundaries (by heading) with the user if not specified.
3. `write_file` each new note with its portion + minimal frontmatter.
4. Replace the source with a stub index of wikilinks to the new notes
   (or delete content only if the user wants the source removed — prefer
   stub over silent delete).
5. Show the resulting paths.

### Merge

1. `read_file` all sources in order.
2. Concatenate with clear `##` headings per source (and source path comment
   or blurb).
3. `write_file` the destination.
4. Only remove/blank sources if the user explicitly requests cleanup.
5. Confirm order and dest path first if ambiguous.

### Restructure

1. Read the note.
2. Propose a new outline (headings only) before writing.
3. On approval, rewrite via `write_file` preserving substance.
4. Do not drop content silently — call out anything removed.

## Rules

- **Confirm** before overwriting multi-note refactors.
- Preserve meaning; cosmetic cleanup is fine.
- No diff/patch tools — always full-file writes.
- Do not touch deny-listed paths.
- Mention that inbound wikilinks may break if basenames change; offer to
  approximate backlinks via `/backlinks` if needed.
