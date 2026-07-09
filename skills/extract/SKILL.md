---
name: extract
description: >
  Extract patterns across vault notes — TODOs, questions, decisions, or
  custom markers — and present or save a structured list. Use when the user
  runs /extract, asks to "find all TODOs", "list open questions", or
  "extract decisions from my notes".
argument-hint: "[todos|questions|decisions|custom] [scope/query]"
---

# Extract Patterns

Scan relevant notes and collect repeated structural patterns.

## Prerequisites

- MCP tools: `search_notes`, `read_file`, optionally `list_objects`, `write_file`

## Arguments

Parse `$ARGUMENTS`:

- Pattern kind: `todos` | `questions` | `decisions` | free-form marker
- Optional scope: folder path or search keywords

Default kind: `todos`.

## Pattern heuristics

| Kind | Look for |
|------|----------|
| todos | `- [ ]`, `TODO`, `FIXME`, `WIP` |
| questions | lines with `?`, `OPEN:`, `Q:` |
| decisions | `Decision:`, `ADR`, `We decided`, `DECIDED` |
| custom | user-provided string/regex-like phrase (literal match) |

## Steps

1. **Scope the corpus**
   - With keywords → `search_notes` then read top hits.
   - With folder → `list_objects` (`types: ["note"]`, `recursive: true`) then
     read files (cap at a reasonable N; prefer search if the folder is huge).
   - Unscoped → search for the marker terms themselves.

2. **Extract**
   - For each note, collect matching lines with path + optional nearby heading
     if obvious from the file text.
   - Deduplicate near-identical items.

3. **Present**
   Group by note or by theme:

   ```markdown
   ## path/to/note.md
   - [ ] item (context)
   ```

4. **Optional save**
   Only if the user asks to save: `write_file` to e.g.
   `Extracts/YYYY-MM-DD-<kind>.md`.

## Rules

- Read-only by default.
- Do not "complete" TODOs in source notes unless asked.
- Be honest about partial coverage when the vault is large.
