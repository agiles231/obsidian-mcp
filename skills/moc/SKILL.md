---
name: moc
description: >
  Generate or update a Map of Content (MOC) note that links related vault
  notes on a topic. Use when the user runs /moc, asks for a "map of content",
  "index note for", or wants a hub note linking a cluster of notes.
argument-hint: "<topic> [output path]"
---

# Map of Content

Build a hub note that organizes related notes with wikilinks.

## Prerequisites

- MCP tools: `search_notes`, `list_objects`, `read_file`, `write_file`

## Arguments

- Topic (required): subject of the MOC
- Optional output path; default `MOCs/<Topic-Slug>.md`

## Steps

1. **Gather related notes**
   - `search_notes` on the topic (`limit` 15–20).
   - Optionally `list_objects` on a relevant folder.

2. **Cluster**
   - Group hits into 3–7 sensible sections (e.g. Specs, Decisions, How-to,
     Open questions).
   - Skim via `read_file` only when the title is ambiguous.

3. **Draft MOC**
   ```markdown
   ---
   title: <Topic> — MOC
   tags: [moc]
   ---

   # <Topic> — MOC

   ## Overview
   <2–3 sentences on what this map covers>

   ## <Cluster name>
   - [[path/without/extension or note name]] — <one-line blurb>
   ```

   Use Obsidian-friendly wikilinks. Prefer vault-relative path links when
   names may collide.

4. **Write or update**
   - If the MOC path exists: `read_file`, merge new links without deleting
     user-curated entries, then `write_file`.
   - If new: `write_file` the full draft.
   - Ask before destructive rewrites of a rich existing MOC.

5. **Report**
   - Path, section counts, and notable gaps (topics searched but not found).

## Rules

- A MOC is an index, not a dump of full note bodies.
- Do not create empty wikilinks to notes that do not exist unless the user
  wants placeholders.
