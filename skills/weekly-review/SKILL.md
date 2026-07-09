---
name: weekly-review
description: >
  Summarize the past week's daily notes into a weekly review note. Use when
  the user runs /weekly-review, asks for a "weekly review", "summarize my
  week", or wants to roll daily notes into a week note.
argument-hint: "[week-start YYYY-MM-DD]"
---

# Weekly Review

Collect this week's daily notes, synthesize themes, and write a weekly note.

## Prerequisites

- MCP tools: `list_objects`, `read_file`, `write_file`, optionally `daily_note` / `search_notes`
- Daily notes live under the vault's configured daily folder (often `Daily/`)

## Arguments

`$ARGUMENTS` may be a week-start date (`YYYY-MM-DD`, Monday preferred).
Default: the Monday of the current week (or last 7 days if preferred and stated).

## Steps

1. **Discover daily notes**
   - `list_objects` on the daily folder (`path: "Daily"` or user convention)
     with `types: ["note"]`, `recursive: true` if needed.
   - Filter names/dates to the target week.

2. **Read the week's dailies**
   - `read_file` each matching note.
   - If a day is missing, note the gap; do not invent entries.

3. **Synthesize**
   Structure the weekly note as:

   ```markdown
   ---
   title: Week of YYYY-MM-DD
   tags: [weekly-review]
   ---

   # Week of YYYY-MM-DD

   ## Highlights
   - ...

   ## Themes
   - ...

   ## Completed
   - ...

   ## Open loops
   - ...

   ## Daily sources
   - [[Daily/YYYY-MM-DD]]
   ```

4. **Write**
   - Default path: `Weekly/YYYY-MM-DD.md` (week start date).
   - `write_file` with full content.
   - Confirm before overwrite if the weekly note already exists (probe with
     `read_file`).

5. **Report**
   - Path written, day coverage, top 3 highlights.

## Rules

- Ground every claim in daily note content.
- Prefer synthesis over dumping raw daily text.
- Do not modify the daily notes themselves.
