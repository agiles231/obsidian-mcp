# 13. Tool design: orthogonal primitives, specific tools only for invariants

- **Status:** Accepted
- **Date:** 2026-07-10

## Context

We are deciding how large the tool surface should be. Candidate tools include
read/write/patch frontmatter, tag/untag, move note/folder, richer search,
projected listing, prepend, and more. Two failure modes pull in opposite
directions:

- **Too many tools:** every tool's schema sits in the model's context
  permanently, and every call forces a "which tool?" decision. Overlapping tools
  (two ways to do one thing) cause wrong picks and waste context.
- **Too few / too general:** a single `write_file` pushes correctness onto the
  model (reformatting, dropped content, token cost) and prevents the server from
  guaranteeing invariants.

We need a rule for admitting a tool.

## Decision

**Favor a small set of orthogonal primitives. Add a specific tool only when it
encodes an invariant the primitive cannot guarantee.**

The admission test for any proposed tool:

> Does it prevent a class of error the model would otherwise make, or is it just
> sugar over an existing primitive? Invariant-carrying tools are worth the
> surface area even if they overlap a primitive. Pure sugar adds selection
> ambiguity without buying safety — reject it.

Two forces, made explicit:
- **Selection clarity** (favors fewer): low overlap so the model always knows
  which tool to use.
- **Error prevention** (favors a few specific): the server guarantees something
  the model would otherwise have to get right by hand and sometimes won't.

### Applying the test (initial rulings)

**Admitted — carry an invariant:**
- `patch_note_frontmatter` — merges frontmatter, preserving body byte-for-byte
  and leaving untouched keys intact. Makes body-loss and dropped-keys
  *impossible*. Foundation-defining.
- `read_note_frontmatter` — returns *parsed, structured* metadata without a body
  round-trip; the model never hand-parses YAML.
- `move_note` / `move_folder` — carry link-integrity handling (advisory per
  ADR-0012); a raw path change cannot.
- `trash_note` — recoverable delete (to `.trash`), never hard delete.
- `create_from_template` — guarantees correct core-template variable substitution.

**Conditional — admit only if canonicalizing:**
- `tag_note` / `remove_tag` — justified *only* if they guarantee tag
  canonicalization (dedupe, nested-tag handling, frontmatter-list vs. inline
  `#tag` placement). If they are merely "patch the tags array," they are sugar
  over `patch_note_frontmatter` — fold them.

**Rejected — sugar over a primitive:**
- `write_note_frontmatter` (replace whole block) — expressible as
  `patch_note_frontmatter` with a replace/`unset` semantic. Two tools for one job.
- `prepend_note` — pure sugar over the append mental model; the real need
  (frontmatter) is covered by `patch_note_frontmatter`.

### Foundations first

Most candidate tools rest on two subsystems. Build these once and the tools fall
out cheaply, so sequence by foundation, not by tool:

1. **Frontmatter engine** (parse ↔ render, body-preserving) → read/patch
   frontmatter, tag tools, search filters, listing projection.
2. **Link graph** (forward links + backlinks, advisory resolution) → backlinks,
   safe-ish move/rename, orphans, MOC.

## Consequences

- **+** A repeatable admission test instead of case-by-case debate.
- **+** Tool count stays moderate; each tool has a clear, non-overlapping purpose.
- **+** Server-enforced invariants (no body loss, no dropped keys) rather than
  hoping the model gets full-file rewrites right.
- **−** Some ergonomic conveniences are deliberately omitted as sugar.
- **−** "Carries an invariant" requires judgment; borderline cases (tag tools)
  need a canonicalization decision before admission.

## Related

- ADR-0009 (write-tool strategy — this generalizes its "fewer tools" rationale),
  ADR-0012 (ownership boundary).
