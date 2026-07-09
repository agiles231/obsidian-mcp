# obsidian-mcp

An [MCP](https://modelcontextprotocol.io) server (stdio / JSON-RPC 2.0) that
lets an LLM agent read and write an [Obsidian](https://obsidian.md) vault via
the local filesystem.

Built in Go on [`mcp-stdio-go`](https://github.com/agiles231/mcp-stdio-go).
No Obsidian plugins, no network listener — stdio only.

## Install

Requires **Go 1.24+**.

```bash
go install github.com/agiles231/obsidian-mcp/cmd/obsidian-mcp@v0.1.1
```

Or install the latest tagged release:

```bash
go install github.com/agiles231/obsidian-mcp/cmd/obsidian-mcp@latest
```

The binary is placed in `$(go env GOPATH)/bin` (ensure that directory is on
your `PATH`).

Pre-built binaries for Linux, macOS, and Windows are attached to
[GitHub Releases](https://github.com/agiles231/obsidian-mcp/releases).

### From source

```bash
git clone https://github.com/agiles231/obsidian-mcp.git
cd obsidian-mcp
go build -o obsidian-mcp ./cmd/obsidian-mcp
```

## Run

```bash
obsidian-mcp \
  --vault my-vault \
  --root /path/to/vault \
  --deny ".obsidian,private" \
  --write-allow "**"
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--vault` | *(required)* | Logical vault name (used in URNs / daily notes) |
| `--root` | *(required)* | Absolute path to the vault directory |
| `--read-allow` | empty = all | Comma-separated read allow globs |
| `--write-allow` | empty = **none** | Comma-separated write allow globs (fail-closed) |
| `--deny` | `.obsidian` | Comma-separated deny globs (deny wins) |

### Claude Code / MCP client config

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "obsidian-mcp",
      "args": [
        "--vault", "my-vault",
        "--root", "/path/to/vault",
        "--deny", ".obsidian,private",
        "--write-allow", "**"
      ]
    }
  }
}
```

Use an absolute path to the binary if it is not on `PATH`.

## Tools

| Tool | Purpose |
|------|---------|
| `read_file` | Read any file content |
| `list_objects` | List notes, folders, canvases, attachments |
| `write_file` | Create/overwrite any file |
| `append_note` | Append to a markdown note |
| `daily_note` | Read/append/create today's daily note |
| `search_notes` | Full-text search (stdlib TF-IDF index) |

Notes are identified with `urn:obsidian:` URNs or bare vault-relative paths.
See [`docs/urn-spec.md`](docs/urn-spec.md).

## Skills

Agent slash-command workflows live under [`skills/`](skills/). See
[`skills/README.md`](skills/README.md) for discovery setup.

## Security posture

- **Direct filesystem access** — no third-party Obsidian plugins
- **stdio only** — no network listener
- **Allow/deny path globs** enforced before any tool I/O
- **Deny ⇒ not-found** for sensitive paths (names never confirmed to the agent)
- Vault boundary via Go `os.Root` containment

Design decisions: [`docs/adr/`](docs/adr/).

## Development

```bash
make check   # fmt, vet, lint, test
make build
make test
```

Pinned dependency: `github.com/agiles231/mcp-stdio-go v0.1.0`.

For local joint development with the framework sibling repo:

```bash
go work init . ../mcp-stdio-go
# go.work is gitignored — do not commit a replace in go.mod
```

## License

MIT — see [LICENSE](LICENSE).
