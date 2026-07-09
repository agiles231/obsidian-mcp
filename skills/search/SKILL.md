---
name: search
description: >
  Full-text search the Obsidian vault and incorporate the best hits into
  conversation context. Use when the user runs /search, asks "search my
  vault", "what did I write about X", "find notes on", or needs vault
  knowledge to answer a question.
argument-hint: "<query>"
---

# Search Vault

Use `search_notes` (stdlib TF-IDF index, AND multi-term queries) to find
relevant notes, then selectively `read_file` the best hits and answer with
citations.

## Prerequisites

- MCP tools: `search_notes`, `read_file` (optional `list_objects` for browsing)
- Vault name for `search_notes`. If unknown, ask once.

## Arguments

`$ARGUMENTS` is the search query. If empty, ask for keywords.

Query tips for this server:

- Space-separated terms → **AND** semantics
- Case-insensitive; no stemming or fuzzy match in v1
- Prefer distinctive keywords over long natural-language questions
- If a prose question is given, extract 2–5 strong terms for the tool query

## Steps

1. **Search**
   ```json
   {
     "vault": "<name>",
     "query": "<terms>",
     "limit": 10
   }
   ```
   Default `limit` 8–10. Raise only if the first pass is thin.

2. **Triage results**
   - Results include path, score, and a short context snippet.
   - Pick up to **3–5** notes that actually answer the question.
   - Skip near-duplicates and low-relevance hits.

3. **Read selectively**
   - Call `read_file` with `ref` set to the path or URN for chosen notes.
   - Do **not** dump the entire vault into context.
   - If a note is huge, skim for the relevant section in your reply; still
     prefer full-file reads (no `read_section` tool yet).

4. **Answer**
   - Synthesize an answer grounded in the notes.
   - Cite paths or URNs for each claim (`path/to/note.md` or
     `urn:obsidian::<vault>:note:path/to/note.md`).
   - Quote sparingly; paraphrase when enough.
   - If nothing relevant: say so, suggest alternate terms, optionally try
     one refined query.

5. **Optional follow-up**
   - Offer to open a specific note, append a finding to the daily note, or
     refine the query.

## Rules

- Search is read-only; never write unless the user asks.
- Deny-listed paths are invisible by design — do not probe for them.
- Do not claim completeness; the index covers `.md` notes only.
- Prefer accurate "I don't see that in the vault" over hallucinated notes.
