# Obsidian-MCP URN Specification

- **Status:** Draft
- **Version:** 0.1.0
- **Last updated:** 2026-06-29

This document defines the identifier format that `obsidian-mcp` uses to name
notes (and, later, other vault resources) across tool arguments, tool outputs,
and — when added — MCP resources. It is the single, canonical identity
vocabulary for the server.

---

## 1. Why a URN (and not a URL)

A **URL locates** — it answers *"where/how do I get this,"* and is permitted to
have side effects. `obsidian://open?vault=X&file=Y` is a real, OS-registered
locator that launches the Obsidian app. A **URN names** — it answers *"what is
this,"* with no locating or action semantics.

Identity is a naming concern, so we use a URN. Reusing `obsidian://` for
identity would be semantic squatting on a live locator scheme with a grammar
Obsidian does not accept, and any such string leaking into a clickable context
(chat UI, a note body, logs) would no-op or misfire. We use the standard
`urn:` form with an `obsidian` namespace rather than inventing a new top-level
scheme.

> A genuine `obsidian://open?...` URL MAY still be emitted separately as a
> convenience "open in Obsidian" affordance. That is a *locator*, kept distinct
> from *identity*. This spec governs identity only.

**The URN wrapper does not make path-based identity stable.** A rename still
invalidates a path-derived URN. What the URN buys is: a clear signal that the
string is a name (not a raw filesystem path), a uniform slot for section
anchors, alignment with MCP resource URIs, and a clean seam for a future
location-independent `id` addressing mode (see §9).

---

## 2. Canonical form

```
urn:obsidian:<user>:<vault>:<type>:<identifier>[#<anchor>]
```

In v1:

- `<user>` is **always empty** (reserved; see §9), producing a literal `::`.
- `<type>` is always `note`.
- `<identifier>` is a vault-relative path.

Example (a note):

```
urn:obsidian::my-vault:note:Projects/obsidian-mcp.md
```

Example (a section within a note):

```
urn:obsidian::my-vault:note:Projects/obsidian-mcp.md#Design#Identity
```

Example (a block within a note):

```
urn:obsidian::my-vault:note:Daily/2026-06-29.md#^a1b2c3
```

Example (a path with characters that must be encoded — see §6):

```
urn:obsidian::my-vault:note:Meeting%20Notes/Q3%20Planning.md
```

---

## 3. Grammar (ABNF)

Per RFC 5234. The `urn:` scheme token and the `obsidian` namespace identifier
are **case-insensitive**; everything after them (user, vault, type, identifier,
anchor) is **case-sensitive**.

```abnf
obsidian-urn = "urn:obsidian:" user ":" vault ":" rtype ":" identifier
               [ "#" anchor ]

user         = *field-char           ; RESERVED — MUST be empty in v1
vault        = 1*field-char          ; non-empty vault name
rtype        = "note"                ; v1; "id" and others reserved (§9)
identifier   = path                  ; when rtype = "note"

path         = segment *( "/" segment )
segment      = 1*field-char          ; no leading/trailing slash; no empty segment

anchor       = heading-path / block-ref
heading-path = heading *( "#" heading )
heading      = 1*field-char
block-ref    = "^" block-id
block-id     = 1*( ALPHA / DIGIT / "-" )

field-char   = literal-char / pct-encoded
literal-char = %x21-7E except ( ":" / "/" / "#" / "%" / "?" )
pct-encoded  = "%" HEXDIG HEXDIG
```

Notes:

- `/` is the path-segment separator and appears **only** inside `path`.
- `:` separates the four fixed NSS fields. A literal `:` inside any field value
  MUST be percent-encoded (`%3A`), so splitting the NSS on literal `:` is
  unambiguous and yields exactly four fields.
- The first `#` introduces the anchor (the URI fragment). Within the anchor,
  `#` separates nested headings — this is a deliberate lift of Obsidian's link
  grammar and a documented superset of strict RFC 3986 fragment syntax. A
  literal `#` inside a heading or filename MUST be percent-encoded (`%23`).
- `?` is reserved (RFC 8141 r-/q-components) and MUST NOT appear unencoded.

---

## 4. Fields

| Field        | v1 value        | Meaning                                              |
|--------------|-----------------|------------------------------------------------------|
| `user`       | empty           | Reserved owner/principal slot (§9). Always `::`.     |
| `vault`      | configured name | Names which vault the identifier belongs to.         |
| `type`       | `note`          | Addressing/resource discriminator.                   |
| `identifier` | path            | For `type=note`: the vault-relative path of the note.|

The `vault` value is the logical vault name from server configuration, **not**
the absolute filesystem root. The mapping from vault name → root lives in
configuration and never appears in an identifier.

---

## 5. Anchors (sections)

The anchor names a location *within* a note, using Obsidian's own grammar:

- **Heading:** `#Heading`
- **Nested heading:** `#Heading#Subheading` (path through the heading tree)
- **Block reference:** `#^blockid` (Obsidian's stable per-block id)

An identifier with no anchor refers to the whole note. Heading anchors share
the note's volatility (renaming a heading breaks `#Heading`); block references
are Obsidian's native stable-within-note mechanism and survive heading edits.

---

## 6. Percent-encoding

Any character in a field value that is not a `literal-char` MUST be
percent-encoded as one or more `%HH` octets of its UTF-8 encoding. In practice
this means encoding, at minimum:

| Character        | Encoded |
|------------------|---------|
| space            | `%20`   |
| `:`              | `%3A`   |
| `#`              | `%23`   |
| `%`              | `%25`   |
| `?`              | `%3F`   |
| non-ASCII (UTF-8)| `%HH…`  |

`/` is **not** encoded inside `path` (it is the segment separator). Producers
MUST emit canonical, minimally-encoded URNs; consumers MUST percent-decode each
field after splitting on the structural delimiters.

---

## 7. Parsing algorithm

1. Verify the prefix `urn:obsidian:` case-insensitively. Reject otherwise.
2. Split off the anchor at the **first** literal `#`. The part before is the
   *NSS body*; the part after (if any) is the raw anchor.
3. Split the NSS body on literal `:` into **exactly four** components:
   `user`, `vault`, `type`, `identifier`. Any other count is an error.
4. Validate:
   - `user` is empty (v1),
   - `vault` is non-empty,
   - `type` is `note` (v1),
   - `identifier` is a non-empty path with no leading/trailing slash and no
     empty segments.
5. Percent-decode `vault` and each path segment.
6. If an anchor is present: if it begins with `^`, it is a `block-ref`;
   otherwise split it on `#` into one or more headings and percent-decode each.

The decoded path is then handed to the vault access-control layer
(`Vault.resolve`) for security validation (containment, symlink, allow/deny)
before any disk access. Identity parsing performs **no** filesystem access and
grants **no** access by itself.

---

## 8. Resolver behavior (liberal in, canonical out)

To keep one vocabulary without burdening callers:

- **Accepted inputs:** a full canonical URN (this spec), **or** a bare
  vault-relative path (literal, unencoded; e.g. `Projects/obsidian-mcp.md`).
  The bare-path form targets the configured default vault. Wikilink/name-based
  input (`[[Note#Heading]]`) is **not** accepted — name resolution introduces
  ambiguity and a failure mode callers would have to special-case.
- **Emitted output:** always the canonical URN of this spec. All tool outputs,
  search hits, and (future) resource URIs use the canonical form.

---

## 9. Reserved extensions (additive, non-breaking)

These are reserved now so future use does not change the grammar:

- **`user` field.** Currently empty. Intended for an owner/principal when the
  server ever addresses more than one. Populating it is additive — the field
  already exists in every identifier.
- **`type = id`.** A future location-independent addressing mode, e.g.
  `urn:obsidian::my-vault:id:01HZX...`, resolved via an id→path index. This is
  the seam for true rename-stability; it slots in as a new `type` branch in the
  resolver without changing existing `note` identifiers.
- **Other `type` values** for non-note resources (attachments, canvases) MAY be
  added the same way.

---

## 10. Non-goals

- This spec does not define MCP *resources*; it defines the identifier they
  will reuse when added.
- It does not provide rename stability for `type=note` (see §1, §9).
- It does not define the `obsidian://open?...` convenience locator.
