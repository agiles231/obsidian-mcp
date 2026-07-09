# 11. Stdlib-only search index

- **Status:** Accepted
- **Date:** 2026-07-08

## Context

The `search_notes` tool requires full-text search across the vault. Several
approaches were considered:

| Approach | Pros | Cons |
|----------|------|------|
| Bleve (Go search lib) | Feature-rich, BM25, analyzers | Transitive deps, network risk |
| SQLite FTS5 | Battle-tested, fast | Still has deps to audit |
| Grep-style scan | No deps | Slow on large vaults, no ranking |
| Stdlib inverted index | No deps, fully auditable | Must build ourselves |

The project's prime directive is data privacy: vault data must never leave the
machine. This extends to **supply chain risk** — any dependency could contain:

- Telemetry or phone-home code
- Network calls in `init()` functions
- Transitive dependencies with unknown behavior

Without auditing every line of Bleve and its dependencies, we cannot guarantee
no network I/O. The audit burden grows with each dependency update.

## Decision

Build a **stdlib-only search index** using only Go's standard library. No
external dependencies for the search subsystem.

### Design

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Tokenizer  │ ──▶ │ Inverted Idx │ ──▶ │   Ranker    │
│ (text→terms)│     │ (term→docs)  │     │  (TF-IDF)   │
└─────────────┘     └──────────────┘     └─────────────┘
```

**Components:**

1. **Tokenizer** — lowercase, split on non-alphanumeric, emit (term, position)
2. **Inverted index** — `map[term][]Posting` where Posting = {path, positions, TF}
3. **Ranker** — TF-IDF scoring (simpler than BM25, sufficient for vault scale)
4. **Query executor** — AND semantics for multi-term queries

**Scope for v1:**

| Feature | Status |
|---------|--------|
| Case-insensitive matching | Included |
| TF-IDF ranking | Included |
| Multi-term AND queries | Included |
| In-memory index | Included |
| Stemming | Deferred |
| Stop words | Deferred |
| Fuzzy matching | Deferred |
| Phrase queries | Deferred |
| Persistent index | Deferred |

**Integration:**

- Index builds on first search or explicit rebuild
- Scans `.md` files only (notes, not attachments)
- Respects vault deny-list during indexing
- Stored in memory (vaults typically <10k notes, <100MB text)

## Rationale

Obsidian vaults are small by search engine standards. A 10,000-note vault with
1KB average per note is ~10MB of text. An in-memory inverted index for this
corpus is trivial — we don't need Lucene-grade infrastructure.

The stdlib-only constraint means:

- **Zero network risk** — we control every syscall
- **Full auditability** — ~300 lines of code to review
- **No supply chain attacks** — no transitive deps
- **Fast builds** — no CGO, pure Go

TF-IDF is "good enough" for knowledge base search. Users searching their own
notes have strong priors — they remember keywords. Fancy ranking matters less
than in web search.

## Consequences

- **+** Guaranteed no network I/O from search subsystem.
- **+** Fully auditable, maintainable in-house.
- **+** No dependency updates to track.
- **+** Fast cold start (no index persistence needed for small vaults).
- **−** No stemming — "running" won't match "run" (acceptable for v1).
- **−** No fuzzy matching — typos won't match (acceptable for v1).
- **−** Must rebuild index on restart (fast for vault-scale data).

## Future considerations

If search quality becomes insufficient:

1. Add stemming via Porter Stemmer (stdlib-implementable, ~100 lines)
2. Add persistence for large vaults (gob encoding to `.obsidian/search.idx`)
3. Revisit external libs with full network audit + seccomp sandboxing
