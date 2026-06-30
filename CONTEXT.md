# obsidian-mcp — Session Handoff / Catch-up

This file is a catch-up mechanism for a fresh session starting work on the
**obsidian-mcp** server. It captures the goal, the hard constraints, the
framework it builds on, the decisions already made, and the open questions.

> The framework (`mcp-stdio-go`) was built in a separate session and is
> considered stable for our needs. This repo (`obsidian-mcp`) is the actual
> product and is **empty except for boilerplate** (`.gitignore`, `LICENSE`,
> `README.md`) — no `go.mod` yet. Starting from scratch here.

---

## 1. The goal

Build an MCP (Model Context Protocol) server for **Obsidian**, in Go, that lets
an LLM/agent read and work with an Obsidian vault — starting with reading a
note. It speaks MCP over **stdio** (JSON-RPC 2.0), via our own framework.

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
  vet third-party plugin code for data exfiltration. Filesystem access (or a
  vetted local Obsidian CLI — see open questions) keeps everything local and
  auditable. **stdio-only is itself a security posture: no network listener,
  nothing to bind or firewall.**

## 3. The framework it builds on: `mcp-stdio-go`

- **Repo:** `github.com/agiles231/mcp-stdio-go`, located locally at
  `~/workspace/mcp-stdio-go` (sibling of this repo).
- **No tag yet.** For tight local iteration, add a replace directive to this
  repo's `go.mod`:
  ```
  replace github.com/agiles231/mcp-stdio-go => ../mcp-stdio-go
  ```
  (Tag `v0.1.0` and pin properly once the API settles.)
- Built from scratch — **no third-party MCP SDK.** Layered packages:
  `protocol` (wire types) ← `transport` (owns stdout, newline-delimited
  JSON-RPC) ← root `mcp` package (public API). **stdout is owned exclusively
  by the transport** — all logging goes to stderr via `slog`. Never `fmt.Print`
  to stdout; it corrupts the wire.

### Public API you implement against (stable surface)

```go
// A tool is this interface. Execute MAY run concurrently (bounded by
// WithMaxConcurrency); you own your own thread-safety. For writes to the
// SAME file, serialize yourself (per-path mutex).
type Tool interface {
    Name() string                 // must match ^[a-zA-Z0-9_-]+$ (no spaces!)
    Description() string
    Schema() InputSchema
    Execute(ctx context.Context, args json.RawMessage) ([]Content, error)
}

// Returning a non-nil error => tool-level error (isError:true) carrying the
// message. It is NOT a JSON-RPC protocol error.

type InputSchema struct {
    Type       string              // "object"
    Properties map[string]Property
    Required   []string
}
type Property struct {
    Type        string  // currently Type + Description only (more is additive)
    Description string
}

// Result content. Helpers: mcp.Text(s), mcp.Image(bytes, mime), mcp.Audio(bytes, mime)
type Content struct { Type, Text, Data, MimeType string }

// OPTIONAL interface — advisory risk hints. Implement it to surface them in
// tools/list. Hints are *bool; use mcp.HintTrue() / mcp.HintFalse().
type Annotated interface { Annotations() Annotations }
type Annotations struct {
    Title           string
    ReadOnlyHint    *bool  // read_note => HintTrue()
    DestructiveHint *bool  // delete/overwrite => HintTrue()
    IdempotentHint  *bool
    OpenWorldHint   *bool  // local vault => HintFalse()
}
```

Server lifecycle:
```go
srv := mcp.NewServer("obsidian-mcp", "0.1.0",
    mcp.WithMaxConcurrency(n),   // optional, default 8
    mcp.WithLogger(logger),      // optional, slog; defaults to slog.Default()
    mcp.WithIO(r, w),            // optional, for tests; defaults to os/stdin/out
)
srv.Register(tool1, tool2, ...)  // panics on nil / empty-name / duplicate
err := srv.Run(ctx)              // blocks; cancel ctx for graceful shutdown
```

### Framework behaviors already handled for you

- Full MCP lifecycle: `initialize` → `notifications/initialized` → operational;
  `ping`; `tools/list`; `tools/call`. Protocol version `2024-11-05`.
- **Capabilities advertised from the registry** (only claims `tools` when tools
  are registered).
- **Bounded concurrent dispatch** with **panic recovery** (a panicking tool
  cannot crash the server).
- **Per-request cancellation**: `notifications/cancelled` cancels that call's
  `context`. Cancellation is **cooperative** — your `Execute` must watch
  `ctx.Done()`/`ctx.Err()` to actually stop work (Go can't kill a goroutine).
  The framework suppresses the response for a cancelled call.
- Tool-level errors (`isError`) vs JSON-RPC protocol errors are distinguished
  centrally; just return an `error` from `Execute` for the former.

## 4. Decisions already made

- **Tools-first; resources deferred.** MCP "resources" (user-curated, browsable
  context by URI) are a good fit for a vault but are **not** in v1. Rationale:
  a tool call is an explicit, individually gate-able action — every note read is
  a discrete event the access layer can allow/deny/log. Resources are more
  "ambient." Revisit later for user-driven "browse & attach a note."
- **No `notifications/*` list-changed / subscribe in v1** (deferred, additive
  when needed).
- Everything remaining on the framework roadmap is **additive** (new optional
  interfaces; no changes to `Tool` or `Execute`). So building obsidian now
  won't be invalidated by later framework work.

## 5. Open design questions to resolve early (in THIS repo)

1. **Vault access: direct filesystem vs. proxy through an Obsidian CLI.**
   - **RESOLVED (2026-06-29): direct filesystem access.** The Obsidian CLI was
     investigated — it does have a non-interactive mode (not TUI-only, which
     was the worry), but its capability surface is too limited to be worth
     binding to an external process and its lifecycle. We'd end up building our
     own indexing regardless. Direct FS gives full control and keeps everything
     local and auditable.
   - *(historical)* *Obsidian CLI proxy:* would have given Obsidian's
     link-resolution + frontmatter index "for free," but at the cost of an
     external dependency. Rejected.

2. **Note/section identity (the "URN" question).**
   - **RESOLVED (2026-06-29): a `urn:obsidian:` URN over path-based identity.**
     Full spec in [`docs/urn-spec.md`](docs/urn-spec.md). Canonical form:
     `urn:obsidian:<user>:<vault>:<type>:<identifier>#<anchor>` — in v1 `user`
     is always empty (reserved, literal `::`), `vault` is populated, `type` is
     `note`, `identifier` is a vault-relative path, anchor uses Obsidian's own
     grammar (`#Heading`, `#Heading#Sub`, `#^blockid`).
   - **Why URN not URL:** identity is a *naming* concern; `obsidian://` is a
     live OS-registered *locator* (opens the app) and reusing it would be
     squatting. (A real `obsidian://open?...` URL may still be emitted as a
     separate "open in Obsidian" convenience — distinct from identity.)
   - **Path-based, not frontmatter-id.** Path is intrinsic, can't desync, needs
     no plugin, and is legible to the LLM. Rename-breakage is transient and
     self-healing for a re-querying agent; the id approach's cost (a minting
     plugin + an id→path index + cache invalidation) is permanent — bad
     asymmetry. The URN does NOT grant rename-stability; it gives us a clean
     seam (`type=id`, reserved) to add a location-independent id mode later,
     additively. A search-by-basename fallback on resolve-miss can recover most
     moves without any index.
   - **Resolver: liberal in, canonical out.** Accepts a full URN or a bare
     vault-relative path; emits the canonical URN everywhere. Wikilink/name
     input (`[[Note]]`) is rejected — name resolution adds ambiguity and a
     failure mode the client would have to special-case, for little gain.
   - Goal achieved: **one identifier vocabulary shared across read (and later
     resource) and write.**

3. **Access-control layer design.** Where/how to enforce the readable/writable
   allow-list + sensitive-dir deny-list, *before* a tool touches disk. Likely a
   shared component all obsidian tools route through (e.g. a vault abstraction
   that resolves + validates a vault-relative path and refuses escapes via
   `..`, symlinks, or deny-listed dirs). This is the security spine.

## 6. First milestone

A **`read_note`** tool:
- Vault-relative path argument.
- Routes through the access-control layer (reject paths outside the vault /
  inside deny-listed dirs / path-traversal attempts).
- Reads from disk, returns note text as `[]mcp.Content{mcp.Text(...)}`.
- Annotated: `ReadOnlyHint: HintTrue()`, `OpenWorldHint: HintFalse()`.
- Tool name must be `read_note` (regex `^[a-zA-Z0-9_-]+$` — no spaces).

Likely the access-control/vault abstraction comes *first* (or alongside), since
`read_note` should never exist without its guard rails.

## 7. Future tool ideas (captured, not committed)

- `get_note_sections` (list a note's heading structure) + `update_section`
  ((note, section) tuple) — reading sections could later also be resources;
  **writes are always tools** (resources are read-only in MCP).
- Search across the vault (will want `ctx` cancellation since it can be long).

## 8. Working style

Collaborative: in the framework sessions, **the user writes the code and the
assistant guides** with snippets + reasoning (the user rejected direct edits in
favor of writing it themselves). Expect the same here unless told otherwise.
