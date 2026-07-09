---
name: save-session
description: >
  Save a conversation summary to the Obsidian vault as a structured note.
  Optionally link it from today's daily note. Use when the user runs
  /save-session, asks to "save this session", "write a session summary",
  "capture this conversation to the vault", or wants to archive decisions
  and work from the current chat.
argument-hint: "[path or title] [optional focus]"
---

# Save Session

Persist a durable summary of the current conversation into the vault via
`obsidian-mcp` tools. Prefer structured markdown the user can find later.

## Prerequisites

- MCP tools: `write_file`, `append_note`, `daily_note` (optional), `list_objects` (optional)
- Know the vault name (from MCP config). If unclear, ask once.

## Arguments

Parse `$ARGUMENTS` (if any):

- **Path** — if it looks like a vault-relative path ending in `.md`, use it
- **Title / focus** — otherwise treat as the note title or summary focus
- Empty — invent a sensible title from the session topic

Default location if no path given: `Sessions/YYYY-MM-DD-short-slug.md`
(use today's date; slug from the title, lowercase, hyphens).

## Steps

1. **Confirm target** (only if ambiguous)
   - Path and title not clear → ask once, then proceed.
   - Do not overwrite an existing important note without confirmation.
     If the path may already exist, call `read_file` first; on success,
     ask before overwrite or pick a new path.

2. **Draft the note**
   Use this template (adapt headings; drop empty sections):

   ```markdown
   ---
   title: <Title>
   date: <YYYY-MM-DD>
   tags: [session]
   ---

   # <Title>

   ## Summary
   <2–5 sentences: what was done and why it matters>

   ## Decisions
   - <decision and rationale>

   ## Work completed
   - <bullet outcomes, not process narration>

   ## Artifacts
   - `<paths, PRs, commands, file names>`

   ## Open questions / next steps
   - <follow-ups>

   ## Links
   - <related notes or URNs if known>
   ```

   Ground the content in this conversation only. Do not invent decisions
   that were not made. Prefer concrete identifiers over vague language.

3. **Write the note**
   Call `write_file`:
   - `ref`: vault-relative path or full URN
     (`urn:obsidian::<vault>:note:<path>`)
   - `content`: full markdown body

4. **Optional daily-note link**
   If the user asked to "also add to daily" / "link from daily", or if
   linking is clearly wanted:
   - Call `daily_note` with `mode: "append"`, `vault: <name>`, and content
     like:
     ```markdown
     - Session: [[Sessions/YYYY-MM-DD-short-slug]] — <one-line summary>
     ```
   - Prefer wikilink-style paths Obsidian resolves; do not require a
     wikilink tool.

5. **Report**
   - Show the path (and URN if returned).
   - One-line summary of what was saved.
   - Mention if the daily note was updated.

## Rules

- Never write into deny-listed areas (e.g. `private/`).
- Never put secrets, tokens, or credentials into the note.
- Full-file write only (`write_file`); do not invent section-patch tools.
- Keep the note scannable: short bullets beat long prose.
