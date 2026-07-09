---
name: canvas
description: >
  Create an Obsidian canvas (.canvas JSON) from conversation context or
  selected notes. Use when the user runs /canvas, asks to "make a canvas",
  "visual map of these notes", or wants a spatial layout of ideas in the vault.
argument-hint: "<output path> [topic or note paths]"
---

# Create Canvas

Obsidian canvases are JSON files (`*.canvas`). Create them with `write_file`
(general write tool — not note-specific).

## Prerequisites

- MCP tools: `write_file`, optionally `search_notes`, `read_file`, `list_objects`
- Write allow-list must permit `.canvas` paths

## Arguments

- Output path ending in `.canvas` (default: `Canvases/<slug>.canvas`)
- Topic and/or note paths to include as nodes

## Canvas JSON shape (minimal valid)

```json
{
  "nodes": [
    {
      "id": "n1",
      "type": "text",
      "text": "Idea",
      "x": 0,
      "y": 0,
      "width": 250,
      "height": 120
    },
    {
      "id": "n2",
      "type": "file",
      "file": "Projects/auth.md",
      "x": 400,
      "y": 0,
      "width": 400,
      "height": 300
    }
  ],
  "edges": [
    {
      "id": "e1",
      "fromNode": "n1",
      "toNode": "n2"
    }
  ]
}
```

Node types commonly used:

- `text` — free text card (`text` field)
- `file` — vault file embed (`file` = vault-relative path)
- `group` — optional visual grouping if needed

## Steps

1. **Collect content**
   - From arguments / conversation: themes, decisions, note paths.
   - Optionally `search_notes` / `read_file` to attach real files as `file` nodes.

2. **Layout**
   - Grid or simple left-to-right flow; keep coordinates integers.
   - Space nodes (~300–400px gaps) so Obsidian opens a readable canvas.
   - Cap nodes (~15–20) unless the user wants a large map.

3. **Edges**
   - Connect only meaningful relationships (depends-on, related, parent).
   - Unique string ids for nodes and edges.

4. **Write**
   - `write_file` with `ref` = canvas path and `content` = pretty-printed JSON.
   - Confirm overwrite if the path exists.

5. **Report**
   - Path, node count, edge count, how to open in Obsidian.

## Rules

- Emit **valid JSON** only — no markdown fences inside the file body.
- Paths in `file` nodes must be vault-relative and real when possible.
- Do not append to canvases with `append_note` (would corrupt JSON).
- Prefer updating via full `write_file` rewrite if revising a canvas.
