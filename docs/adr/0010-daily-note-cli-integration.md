# 10. Daily note: Obsidian CLI with filesystem fallback

- **Status:** Accepted
- **Date:** 2026-07-07

## Context

The `daily_note` tool must resolve today's date to a vault path, then read or
append content. Obsidian's daily note configuration includes:

- **Folder** — e.g., `Daily/`
- **Date format** — e.g., `YYYY-MM-DD` (Moment.js syntax)
- **Template** — applied on creation (core Templates or Templater plugin)

The configuration lives in `.obsidian/daily-notes.json`. However, templates
introduce complexity:

| Approach | Tradeoff |
|----------|----------|
| Skip templates | Bare file creation; inconsistent with user's vault style |
| LLM fills templates | Non-deterministic, wastes tokens, can't handle Templater expressions |
| Reimplement template engine | Scope creep, will drift from Obsidian |

ADR-0001 chose direct filesystem access to avoid third-party plugin code. But
the concern there was *untrusted* code with potential for data exfiltration.
The Obsidian application itself is trusted — the user installed it. The CLI
(`obsidian://` URI scheme) runs locally and doesn't transmit vault data.

## Decision

Use a **layered approach**:

1. **Creation** — invoke `obsidian://daily` URI to create today's note with
   proper template application. This requires Obsidian to be installed and
   running.

2. **Fallback** — if the URI fails (Obsidian not running, no handler), parse
   `.obsidian/daily-notes.json`, resolve the path from `folder` + `format`,
   and create a bare file. Better than nothing.

3. **Read/append** — once the path is known (via URI success, fallback, or
   file already exists), use existing `read_file` / `append_note` methods.
   No CLI needed for these operations.

The tool exposes a `mode` parameter: `read`, `append`, or `create`. The `create`
mode is where CLI integration matters. Read and append only need path resolution.

## Rationale

Trying to maintain filesystem-only purity for daily notes gains nothing:

- Templates can't be faithfully applied without Obsidian
- Templater plugin syntax (`<% tp.* %>`) is a full scripting language
- Plugins may have side effects on note creation we can't replicate

The CLI is official Obsidian tooling. Using it for *creation* gives full
fidelity. The fallback ensures the tool remains useful even without Obsidian
running — it just won't apply templates.

This is a **limited exception** to ADR-0001, scoped to daily note creation only.
All other file operations remain direct filesystem access.

## Consequences

- **+** Templates and plugins work correctly when Obsidian is running.
- **+** Graceful degradation when Obsidian is unavailable.
- **+** Read/append operations remain pure filesystem (no CLI needed).
- **−** Creation fidelity depends on Obsidian running.
- **−** Introduces optional dependency on Obsidian URI handler.

## Implementation notes

Daily notes config location: `.obsidian/daily-notes.json`

```json
{
  "folder": "Daily",
  "format": "YYYY-MM-DD",
  "template": "Templates/Daily"
}
```

URI to open/create today's daily note:
```
obsidian://daily?vault=<vault-name>
```

Path resolution fallback (Go pseudocode):
```go
func resolveDailyPath(cfg DailyNotesConfig, date time.Time) string {
    formatted := formatMomentDate(cfg.Format, date)
    return filepath.Join(cfg.Folder, formatted+".md")
}
```

Moment.js date format tokens (`YYYY`, `MM`, `DD`, etc.) will need a Go
implementation or lookup table for common patterns.

## Date format escape syntax

Moment.js uses square brackets `[...]` to escape literal text within date
format strings. Text inside brackets is output verbatim, bypassing token
interpretation.

Example:
```
[daily]-YYYY-MM-DD  →  daily-2026-07-07
```

This escape syntax is necessary because Moment.js interprets single-letter
tokens aggressively. Without escaping, literal text containing token characters
would be corrupted:

| Format | Intended | Actual (unescaped) |
|--------|----------|-------------------|
| `daily-YYYY-MM-DD` | `daily-2026-07-07` | `d07mily-2026-07-07` |

The `a` in "daily" matches the am/pm token, producing unexpected output. The
bracket syntax solves this by marking `daily` as literal text to preserve.
