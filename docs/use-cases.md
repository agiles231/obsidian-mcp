# Use Cases and Workflows

This document captures the target workflows for `obsidian-mcp` — how an LLM
agent would use the vault in practice.

---

## 1. Save session as note

**Scenario:** User finishes a coding session with Claude and wants to capture
the conversation summary, decisions made, or generated artifacts as a note.

**Flow:**
1. User: "Save a summary of this session to my vault"
2. Agent summarizes the session
3. Agent calls `write_note` with path and content
4. (Optional) Agent appends link to daily note

**Tools needed:** `write_note`, `append_note`

**Skill opportunity:** A `/save-session` skill could standardize this with
templates, auto-generated titles, and linking.

---

## 2. Research assist

**Scenario:** While working on a task, the agent references the user's existing
notes for context — project specs, prior decisions, architecture docs.

**Flow:**
1. User: "Check my notes on the auth system before implementing"
2. Agent calls `list_notes` or `search_notes` to find relevant notes
3. Agent calls `read_note` to retrieve content
4. Agent incorporates context into its work

**Tools needed:** `read_note` ✓, `list_notes`, `search_notes`

---

## 3. Knowledge query

**Scenario:** User asks a question that their vault can answer — "What did I
decide about X?" or "Summarize my meeting notes from last week."

**Flow:**
1. User asks a question
2. Agent searches vault for relevant notes
3. Agent reads and synthesizes content
4. Agent answers with citations (URNs)

**Tools needed:** `read_note` ✓, `search_notes`

---

## 4. Note discovery / browsing

**Scenario:** User wants to know what's in a folder or find a note they
remember but can't name precisely.

**Flow:**
1. User: "What's in my Projects folder?"
2. Agent calls `list_notes` with path filter
3. Agent presents list with titles/paths

**Tools needed:** `list_notes`

---

## 5. Daily notes integration

**Scenario:** User wants to append findings, TODOs, or session links to their
daily note (a common Obsidian pattern).

**Flow:**
1. User: "Add this TODO to today's daily note"
2. Agent resolves today's date → daily note path (e.g., `Daily/2026-07-05.md`)
3. Agent calls `append_note` with content

**Tools needed:** `append_note`

**Note:** May need date-aware path resolution or a convention for daily note
location.

---

## 6. Note maintenance / section updates

**Scenario:** User wants to update a specific section of an existing note —
change a status, add to a list, update a table.

**Flow:**
1. User: "Update the status in my project tracker"
2. Agent calls `read_note` to get current content
3. Agent identifies the section to modify
4. Agent calls `update_section` or rewrites via `write_note`

**Tools needed:** `read_note` ✓, `update_section` (or `write_note`)

**Note:** Section-level updates require anchor support in the URN and
section-aware editing.

---

## Tool priority matrix

| Tool | Workflows enabled | Complexity | Priority |
|------|-------------------|------------|----------|
| `read_note` | 2, 3, 6 | Done ✓ | — |
| `list_notes` | 2, 4 | Low | High |
| `write_note` | 1 | Medium | High |
| `append_note` | 1, 5 | Low | High |
| `search_notes` | 2, 3 | Medium-High | Medium |
| `update_section` | 6 | High | Low (defer) |

---

## Future considerations

- **Templates:** Predefined note structures for session summaries, meeting
  notes, etc.
- **Wikilink resolution:** Allow `[[Note]]` input for convenience (currently
  rejected per ADR-0002).
- **Frontmatter parsing:** Extract/filter by YAML frontmatter fields.
- **Backlink discovery:** "What links to this note?"
