---
name: frontmatter
description: >
  View, edit, or standardize YAML frontmatter on vault notes. Use when the
  user runs /frontmatter, asks to "add tags to", "set frontmatter",
  "standardize metadata", or bulk-fix note YAML headers.
argument-hint: "[show|set|standardize] <path or query> [key=value...]"
---

# Frontmatter

There is no dedicated frontmatter API yet. Edit frontmatter by reading the
full note and writing it back with an updated YAML block.

## Prerequisites

- MCP tools: `read_file`, `write_file`, `search_notes` / `list_objects` for scope

## Modes

### show

1. `read_file` the note.
2. Display the YAML between leading `---` fences (or say none present).

### set

1. `read_file`.
2. Parse existing frontmatter (best-effort).
3. Apply `key=value` updates from arguments (add block if missing).
4. `write_file` the full note with body preserved exactly.
5. Show a before/after of the YAML only.

### standardize

1. Define the target schema with the user (required keys, tag conventions).
2. Select notes via path, folder list, or search.
3. For each note: read → ensure keys → write if changed.
4. Summarize: updated / skipped / failed counts.
5. Cap bulk updates; confirm if touching more than ~10 notes.

## YAML rules

- Keep frontmatter valid YAML when possible.
- Preserve unknown keys unless asked to strip.
- Do not reorder the body; only the header block should change.
- Quoted strings for values with special characters.
- Tags: prefer list form `tags: [a, b]` unless the vault standard differs.

## Rules

- Full-file rewrite only; never corrupt the body.
- Confirm bulk standardize before running.
- Skip binary/attachment files; notes only (`.md`).
