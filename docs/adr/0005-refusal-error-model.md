# 5. Refusal error model (split taxonomy, deny ⇒ not-found)

- **Status:** Accepted
- **Date:** 2026-06-29

## Context

When the `Vault` refuses an access, what it tells the agent is itself a privacy
decision. The agent's context is a **cloud egress channel** — anything the
agent sees has been sent to a cloud LLM provider — and the prime directive says
vault data, *including filenames*, must not leave without permission. So
confirming that `private/divorce-settlement.md` exists (even while refusing to
read it) would itself transmit a sensitive name off the machine.

But collapsing *every* refusal into one opaque error hurts usability: the agent
can't tell a typo from an out-of-scope path from a secret, making ordinary
mistakes hard to recover from. The threat is prompt-injection + cloud egress,
not a malicious agent — so non-sensitive refusals can safely be explicit.

## Decision

Adopt a **split taxonomy**. Allow/deny are evaluated on the path *string before
any disk access*, in this order — **deny first, deny wins**:

1. **Invalid identifier** *(distinguishable)* — unparseable/absolute input.
2. **Outside the vault** *(distinguishable)* — escapes the root; reveals only
   the (already-known) vault boundary, and `os.Root` fails identically whether
   the external target exists or not.
3. **Not in the readable/writable set** *(distinguishable)* — a valid in-vault
   path not matched by the allow-list. This is a statement about *policy/scope*,
   not file existence, so it is safe (and helpful) to reveal.
4. **Not found ⇔ Denied** *(collapsed, opaque)* — a deny-list match returns the
   exact same result as a genuinely missing file. The agent cannot distinguish
   them, so deny-listed names are never confirmed.

## Consequences

- **+** Sensitive names inside deny-listed areas are never confirmed to the
  (cloud-bound) agent.
- **+** Good ergonomics for non-sensitive refusals — the agent can fix typos and
  understand its own permission scope.
- **−** A minor timing side-channel (deny returns before disk I/O; a real miss
  returns after a stat). Accepted as a non-issue for an LLM client on a local
  machine.
- Full detail (including the denied/resolved path) is logged to **stderr** for
  the local operator; only the opaque message crosses back to the agent.
