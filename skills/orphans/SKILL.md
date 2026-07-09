---
name: orphans
description: >
  Find notes that appear unlinked or weakly connected by sampling the vault
  and checking for inbound wikilink mentions. Use when the user runs
  /orphans, asks for "orphan notes", "unlinked notes", or "notes nothing
  points to". Approximation until native backlink/orphan tooling exists.
argument-hint: "[folder] [limit]"
---

# Orphan Notes (approximation)

Identify notes with no detected inbound links. Best-effort: not a full graph
index.

## Prerequisites

- MCP tools: `list_objects`, `search_notes`, `read_file`

## Arguments

- Optional folder scope (default: vault-wide sample)
- Optional max notes to analyze (default: 50)

## Steps

1. **List candidates**
   - `list_objects` with `types: ["note"]`, `recursive: true`.
   - If the vault is large, restrict to a folder or cap the set and say so.

2. **For each candidate note** (up to limit)
   - Search for its basename via `search_notes`.
   - Spot-check hits with `read_file` for real `[[links]]` to this note.
   - Treat as orphan if no verified inbound links from other notes.

3. **Report**
   ```markdown
   Possible orphans (N scanned, M suspects):
   - path/a.md
   - path/b.md
   ```
   Sort by path. Note false-positive risk (unique titles, path-only links,
   embeds, canvas-only references).

4. **Optional next steps**
   - Offer to create a MOC, add links from a hub note, or open a specific
     orphan for triage.

## Rules

- Read-only.
- Never claim exhaustive graph accuracy.
- Prefer folder-scoped runs for large vaults.
