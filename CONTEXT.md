# obsidian-mcp — Session Handoff / Catch-up

This file is a catch-up mechanism for a fresh session starting work on the
**obsidian-mcp** server. It captures the goal, the hard constraints, the
framework it builds on, the decisions already made, and the current status.

---

## 1. The goal

Build an MCP (Model Context Protocol) server for **Obsidian**, in Go, that lets
an LLM/agent read and work with an Obsidian vault. It speaks MCP over **stdio**
(JSON-RPC 2.0), via our own framework.

## 2. Hard constraints (these drive every design choice)

- **Data privacy is paramount.** The user is extremely cautious: *vault data
  must never be sent over the internet without explicit permission.* This is
  the single most important constraint.
- **Comprehensive access controls are required**, designed into the framework
  layer of the server and enforced **before any tool runs**. Concretely:
  allow-list of readable/writable paths, and a deny-list for sensitive
  directories (e.g. a `private/` folder). Path restrictions are a first-class
  feature, not an afterthought.
- **Why filesystem, not the Obsidian Local REST API plugin:** the user won't
  vet third-party plugin code for data exfiltration. Filesystem access keeps
  everything local and auditable. **stdio-only is itself a security posture:
  no network listener, nothing to bind or firewall.**

## 3. The framework it builds on: `mcp-stdio-go`

- **Repo:** `github.com/agiles231/mcp-stdio-go`, located locally at
  `~/workspace/mcp-stdio-go` (sibling of this repo).
- For tight local iteration, a replace directive in `go.mod`:
  ```
  replace github.com/agiles231/mcp-stdio-go => ../mcp-stdio-go
  ```
- Built from scratch — **no third-party MCP SDK.** Layered packages:
  `protocol` (wire types) ← `transport` (owns stdout, newline-delimited
  JSON-RPC) ← root `mcp` package (public API). **stdout is owned exclusively
  by the transport** — all logging goes to stderr via `slog`.

### Public API (stable surface)

```go
type Tool interface {
    Name() string
    Description() string
    Schema() InputSchema
    Execute(ctx context.Context, args json.RawMessage) ([]Content, error)
}

type Annotated interface { Annotations() Annotations }
type Annotations struct {
    Title           string
    ReadOnlyHint    *bool
    DestructiveHint *bool
    IdempotentHint  *bool
    OpenWorldHint   *bool
}
```

Server lifecycle:
```go
srv := mcp.NewServer("obsidian-mcp", "0.1.0",
    mcp.WithLogger(logger),
)
srv.Register(tool1, tool2, ...)
err := srv.Run(ctx)
```

## 4. Architecture decisions

See [`docs/adr/`](docs/adr/) for full details:

| ADR | Decision |
|-----|----------|
| [0001](docs/adr/0001-direct-filesystem-vault-access.md) | Direct filesystem access (no plugin/CLI) |
| [0002](docs/adr/0002-urn-path-identity.md) | Path-based identity via `urn:obsidian:` URN |
| [0003](docs/adr/0003-vault-sole-disk-gateway.md) | Single `Vault` type as sole disk gateway |
| [0004](docs/adr/0004-os-root-containment.md) | Containment via `os.Root` (Go 1.24+) |
| [0005](docs/adr/0005-refusal-error-model.md) | Split refusal taxonomy; deny ⇒ opaque not-found |
| [0006](docs/adr/0006-threat-model-security-boundaries.md) | Threat model: defense-in-depth, not sandbox |
| [0007](docs/adr/0007-list-objects-unified-listing.md) | Unified `list_objects` tool for vault discovery |
| [0008](docs/adr/0008-object-type-taxonomy.md) | Object type taxonomy (note, folder, canvas, attachment) |
| [0009](docs/adr/0009-write-tool-strategy.md) | Write tool strategy: general write_file, no diffs |
| [0010](docs/adr/0010-daily-note-cli-integration.md) | Daily note: Obsidian CLI with filesystem fallback |
| [0011](docs/adr/0011-stdlib-only-search.md) | Stdlib-only search index (no external deps) |

**URN format:** `urn:obsidian:<user>:<vault>:<type>:<identifier>#<anchor>`
- In v1: `user` empty, `type` is `note`/`folder`/`canvas`/`attachment` (ADR-0008)
- `identifier` is vault-relative path
- See [`docs/urn-spec.md`](docs/urn-spec.md) for full spec

## 5. Current status

### Implemented ✓

- **`internal/vault/`** — Vault abstraction with allow/deny glob matching,
  `os.Root` containment, symlink re-validation; daily note helpers (Moment.js
  date formatting)
- **`internal/urn/`** — URN parser/resolver (liberal in, canonical out); generic
  `Ref` type supports all object types (note, folder, canvas, attachment)
- **`internal/tools/read_file.go`** — Read file content
- **`internal/tools/list_objects.go`** — List vault objects with type filters
- **`internal/tools/write_file.go`** — Create/overwrite files
- **`internal/tools/append_note.go`** — Append to markdown notes
- **`internal/tools/daily_note.go`** — Daily note with Obsidian CLI fallback (ADR-0010)
- **`cmd/obsidian-mcp/`** — CLI entry point with flag-based config
- **Vault registry** — supports multiple vaults (single-vault for now)

### Next steps

See [`TODO.md`](TODO.md) for the full task list.

## 6. Roadmap: tools and use cases

See [`docs/use-cases.md`](docs/use-cases.md) for detailed workflows.

**Implemented tools:**

| Tool | Purpose |
|------|---------|
| `read_file` | Read any file content |
| `list_objects` | Unified vault listing with type filters (ADR-0007) |
| `write_file` | Create/overwrite any file (ADR-0009) |
| `append_note` | Append to markdown notes |
| `daily_note` | Read/append to today's daily note (ADR-0010) |

**Next tools:**

| Tool | Purpose | Enables |
|------|---------|---------|
| `search_notes` | Full-text search | Knowledge queries |

**Planned skills (slash commands):**

Skills are templated workflows exposed as slash commands in Claude Code.

| Skill | Purpose |
|-------|---------|
| `/save-session` | Save conversation summary to vault with templates |
| `/daily` | Quick daily note operations |
| `/search` | Search vault and incorporate context |

## 7. Working style

Collaborative: **the user writes the code and the assistant guides** with
snippets + reasoning. The assistant does not use Edit/Write tools on source
files directly.

## 8. Running the server

```bash
go build -o obsidian-mcp ./cmd/obsidian-mcp

obsidian-mcp \
  --vault my-vault \
  --root /path/to/vault \
  --deny ".obsidian,private"
```

Claude Code config (`.claude/settings.json`):
```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obsidian-mcp",
      "args": ["--vault", "my-vault", "--root", "/path/to/vault"]
    }
  }
}
```
