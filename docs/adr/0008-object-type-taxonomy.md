# 8. Object type taxonomy for vault listing

- **Status:** Accepted
- **Date:** 2026-07-06

## Context

The `list_objects` tool (ADR-0007) needs to classify vault entries by type.
Obsidian supports several object types:

| Object | Storage | Format |
|--------|---------|--------|
| Notes | `.md` files | Markdown |
| Folders | directories | — |
| Canvas | `.canvas` files | JSON (nodes, edges, spatial layout) |
| Bases | `.base` files | JSON (query config over notes) |
| Attachments | images, PDFs, etc. | Binary |

The question: which types deserve first-class classification?

## Decision

Support four object types:

| Type | Matches | Rationale |
|------|---------|-----------|
| `note` | `.md` | Core use case — readable, writable, contains knowledge |
| `folder` | directories | Structural discovery |
| `canvas` | `.canvas` | Novel data structure (spatial layout, connections) not derivable from notes |
| `attachment` | everything else | Catch-all for binary/non-parseable files |

**Bases are intentionally excluded** as a distinct type. A Base is a *view* over
existing notes — it queries frontmatter and tags to produce a table display. The
underlying data lives in the notes themselves. An LLM with access to:

1. `list_objects` (recursive) — discover all notes
2. `read_note` — retrieve content and frontmatter

...can replicate Base functionality by aggregating in-context. The `.base` file
contains query configuration, not data — it provides no value beyond what
recursive listing and note reading already enable.

Canvas, by contrast, *is* the data. A canvas defines nodes, edges, and spatial
relationships that don't exist elsewhere in the vault. It warrants its own type.

## Consequences

- **+** Simple, stable type taxonomy (4 types).
- **+** No special handling for features that are views over existing data.
- **+** Extensible — new types can be added if Obsidian introduces genuinely
  novel object kinds.
- **−** `.base` files appear as `attachment` type. An LLM could read them (JSON)
  but won't know they're query configs without inspecting content.
- **−** If Bases evolve to store data (not just queries), this decision may need
  revisiting.

## Classification logic

```go
func classifyEntry(e os.DirEntry, name string) string {
    if e.IsDir() {
        return "folder"
    }
    switch {
    case strings.HasSuffix(name, ".md"):
        return "note"
    case strings.HasSuffix(name, ".canvas"):
        return "canvas"
    default:
        return "attachment"
    }
}
```
