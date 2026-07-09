---
name: daily
description: >
  Read, create, or append to today's Obsidian daily note via the daily_note
  MCP tool. Use when the user runs /daily, says "today's note", "add to my
  daily note", "what's on my daily", "log this for today", or wants quick
  daily-note operations without picking a path.
argument-hint: "[read|append|create] [content...]"
---

# Daily Note

Operate on **today's** daily note. Path resolution (folder + Moment.js format
from `.obsidian/daily-notes.json`) is handled by the server — do not guess
`Daily/YYYY-MM-DD.md` yourself unless the tool fails and you must fall back.

## Prerequisites

- MCP tool: `daily_note` (`mode`: `read` | `append` | `create`)
- Vault name required by the tool. If unknown, ask once.

## Arguments

Parse `$ARGUMENTS`:

| Input | Mode | Notes |
|-------|------|-------|
| empty or `read` | `read` | Show today's note |
| `create` | `create` | Ensure note exists (Obsidian CLI + template when possible) |
| `append …` or free text | `append` | Everything after `append` (or the whole arg if not a mode word) is content |
| `add …` / `log …` | `append` | Treat rest as content |

If mode is ambiguous and content is present → **append**.
If no content and no mode → **read**.

## Steps

### Read

1. Call `daily_note` with `{ "vault": "<name>", "mode": "read" }`.
2. Present the content (or say the note is missing).
3. If missing, offer to `create` then retry.

### Create

1. Call `daily_note` with `{ "vault": "<name>", "mode": "create" }`.
2. Report whether it was created via Obsidian (templates applied) or bare
   filesystem fallback.
3. Optionally `read` afterward if the user wants the body.

### Append

1. If content is empty, ask what to append — do not append a blank block.
2. Format content as markdown suitable for a daily log, e.g.:
   ```markdown
   ### HH:MM
   <user content>
   ```
   Use local time if helpful; keep the user's wording when they provided
   finished prose.
3. Call `daily_note` with:
   ```json
   { "vault": "<name>", "mode": "append", "content": "<markdown>" }
   ```
4. Confirm success with a short paraphrase of what was added.

## Rules

- Prefer `daily_note` over manual path construction + `append_note`.
- Do not rewrite the whole daily note with `write_file` unless the user
  explicitly wants a full replace.
- Keep appends additive; never delete prior daily content.
- Respect privacy: do not echo deny-listed or secret material.
