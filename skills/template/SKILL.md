---
name: template
description: >
  Create a new vault note from a template note with simple variable
  substitution. Use when the user runs /template, asks to "create from
  template", "new note using", or wants a standardized note structure.
argument-hint: "<template path> <new path> [key=value...]"
---

# Create from Template

Copy a template note to a new path, replacing simple `{{variables}}`.

## Prerequisites

- MCP tools: `read_file`, `write_file`, optionally `list_objects`

## Arguments

Expected shape:

```
<template-path> <dest-path> [key=value ...]
```

Common variables (auto-fill if not provided):

| Variable | Default |
|----------|---------|
| `{{title}}` | derived from dest filename |
| `{{date}}` | today `YYYY-MM-DD` |
| `{{time}}` | local `HH:MM` |
| `{{vault}}` | configured vault name if known |

## Steps

1. **Parse arguments**
   - If template or dest missing, ask.
   - Default template folder guess: `Templates/` — list it if the user only
     names a template title.

2. **Load template**
   - `read_file` on the template path.
   - Fail clearly if missing.

3. **Substitute**
   - Replace `{{key}}` with provided or default values.
   - Leave unknown `{{placeholders}}` intact and warn, or ask — do not
     silently delete them.
   - This is **not** Templater (`<% tp.* %>`). Do not execute scripts.
     If the template contains Templater syntax, tell the user those
     expressions will remain literal unless they create the note via
     Obsidian (e.g. daily note create flow).

4. **Write**
   - Refuse to overwrite without confirmation (`read_file` probe).
   - `write_file` to dest with full content.

5. **Report**
   - New path, variables applied, any leftover placeholders.

## Rules

- Never run template scripts or shell from template bodies.
- Prefer vault templates the user already maintains.
