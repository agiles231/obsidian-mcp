# 9. Write tool strategy: full file, no diffs, no sections

- **Status:** Accepted
- **Date:** 2026-07-06

## Context

We need to decide how the LLM will write/modify notes. Several approaches exist:

| Approach | Description |
|----------|-------------|
| Full overwrite | `write_note(path, content)` — replace entire file |
| Section write | `write_section(path, anchor, content)` — replace one heading's content |
| Diff/patch | Apply unified diff or string replacement |
| Append | `append_note(path, content)` — add to end |

Each has tradeoffs around complexity, reliability, and LLM capability.

## Decision

For v1, provide only:

- **`write_note`** — full file overwrite
- **`append_note`** — append to end of file

**Section operations are deferred.** If `write_section` were useful, symmetric
`read_section` and `get_sections` tools would also be needed — otherwise the LLM
must mentally parse markdown to find anchors, then surgically write. That's an
awkward asymmetry. Either provide the full trio or none. For v1, none.

**Diff-based editing is rejected.** LLMs struggle with diff formats:

- Unified diffs require accurate line numbers — LLMs hallucinate these
- Context lines must match exactly — fragile after any file change
- Whitespace/indentation errors cause patch failures

String replacement (find exact text → replace) is more LLM-friendly than unified
diff, but still adds complexity. For small markdown notes, full overwrite is
simpler and equally effective.

## Rationale

Obsidian notes are typically small (few KB). Full read/write is cheap. The
primary write use cases are:

1. **Save session summary** → `write_note` (new file or overwrite)
2. **Add to daily note** → `append_note`
3. **Update project tracker** → `read_note` + `write_note` (full cycle)

These don't require surgical precision. If a note is too large for full
read/write, it likely should be split into multiple atomic notes (Obsidian
philosophy).

## Consequences

- **+** Simple, reliable write operations.
- **+** No merge conflicts or stale anchor issues.
- **+** Symmetric API — read and write both operate on full files.
- **−** Inefficient for small edits to large files.
- **−** Can't target a specific section without rewriting the whole file.

## Future considerations

If section-level operations become necessary, add as a coherent set:

- `get_sections(path)` → return heading outline (structure only)
- `read_section(path, anchor)` → read content under a heading
- `write_section(path, anchor, content)` → replace content under a heading

Don't add one without the others.
