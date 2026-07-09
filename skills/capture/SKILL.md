---
name: capture
description: >
  Quick-capture a thought, TODO, or snippet into an inbox note in the
  Obsidian vault. Use when the user runs /capture, says "inbox this",
  "quick capture", "park this thought", or wants a low-friction dump
  without filing a full session note.
argument-hint: "<thought or note body>"
---

# Capture

Append a short capture to a fixed inbox note so ideas are not lost.

## Prerequisites

- MCP tools: `append_note` (fallback: `write_file` if creating inbox)
- Default inbox path: `Inbox.md` (override if user names another path)

## Arguments

`$ARGUMENTS` is the capture body. If empty, ask for the thought in one line.

## Steps

1. **Normalize content**
   Format as a dated bullet (unless the user already provided finished markdown):
   ```markdown
   - YYYY-MM-DD HH:MM — <capture text>
   ```

2. **Append**
   Call `append_note`:
   ```json
   {
     "ref": "Inbox.md",
     "content": "\n- YYYY-MM-DD HH:MM — <text>\n"
   }
   ```
   Include a leading newline so captures separate cleanly.
   `append_note` creates the file if missing.

3. **Confirm**
   One line: captured to `Inbox.md` (or chosen path), with a short echo of
   the text.

## Rules

- Do not reorganize the inbox unless asked.
- Do not promote captures into project notes unless the user requests filing.
- Keep it fast — no multi-step confirmation for ordinary captures.
