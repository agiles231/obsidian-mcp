---
name: backlinks
description: >
  Find notes that link to a given note by scanning for wikilinks and path
  mentions. Use when the user runs /backlinks, asks "what links to this",
  "who references", or wants inbound links for a note. (Approximation —
  native backlink index not yet in the server.)
argument-hint: "<note path or name>"
---

# Backlinks (approximation)

There is no dedicated backlink tool yet. Approximate by searching for the
note's name and path fragments, then verifying links in file bodies.

## Prerequisites

- MCP tools: `search_notes`, `read_file`, optionally `list_objects`

## Arguments

`$ARGUMENTS` = target note path or basename (e.g. `Projects/auth.md` or `auth`).

## Steps

1. **Normalize the target**
   - Resolve to a vault-relative path if possible (`list_objects` / user input).
   - Derive search tokens:
     - basename without extension
     - full path
     - path without `.md`

2. **Search**
   - `search_notes` for the basename (and path tokens if useful).
   - `limit` 20.

3. **Verify**
   - `read_file` candidate notes.
   - Count as a backlink only if the body contains something like:
     - `[[basename]]` / `[[path]]` / `[[basename|alias]]`
     - markdown links to the path
     - explicit URN for the note
   - Exclude the target note itself.

4. **Report**
   ```markdown
   Backlinks to <target>:
   - path/a.md — link form: [[...]]
   - path/b.md — link form: ...
   ```
   If none: say so; mention this is a best-effort scan, not Obsidian's
   backlink engine.

## Rules

- Do not modify notes.
- Prefer false negatives over false positives when unsure.
- Note limitation: rename-unstable; basename collisions possible.
