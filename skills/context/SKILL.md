---
name: context
description: >
  Load relevant Obsidian notes into the conversation as working context.
  Use when the user runs /context, says "load my notes on", "pull in vault
  context for", "before we start, check my notes about", or wants the agent
  grounded in existing vault material.
argument-hint: "<topic or keywords>"
---

# Load Context

Search and read vault notes that matter for the current task, then keep them
as working background for subsequent reasoning.

## Prerequisites

- MCP tools: `search_notes`, `read_file`, optionally `list_objects`

## Arguments

`$ARGUMENTS` is the topic or keywords. If empty, infer from recent conversation
or ask.

## Steps

1. **Find candidates**
   - Call `search_notes` with strong keywords and `limit` 10.
   - If the topic is a folder name, also `list_objects` with that `path` and
     `types: ["note"]`.

2. **Select**
   - Choose 2–5 notes that best match the topic.
   - Prefer specs, ADRs, decisions, and project trackers over random hits.

3. **Read**
   - `read_file` each selected note.
   - Mentally retain structure, decisions, constraints, and open issues.

4. **Brief the user**
   Present a short context pack:
   - Bullet list of loaded notes (path + one-line why)
   - Key constraints / decisions that must not be violated
   - Gaps: what the vault does *not* cover

5. **Continue**
   Use this context for subsequent work in the session without re-reading
   unless files may have changed.

## Rules

- Do not write to the vault unless asked.
- Do not flood context with low-score search hits.
- Cite paths when relying on vault facts later in the session.
