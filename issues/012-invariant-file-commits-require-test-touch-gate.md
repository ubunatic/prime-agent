# Issue 012: Require a Test-File Touch (or Explicit Waiver) When Committing to Invariant Files

**Status:** Open

---

## Problem

Same root cause as [`011`](011-fork-sync-verify-test-debt-detection.md) and
[`010`](010-upstream-v0.8.0-features-and-improvements.md): fork commits that intentionally
change behavior in an invariant-covered file (see `defaultInvariantRules` in
[`tools/fork-sync/invariants.go`](../tools/fork-sync/invariants.go)) can land without touching
every test that asserted the old behavior, and nothing in the commit workflow itself prompts a
check. `011` proposes catching this after the fact via `fork-sync verify`; this issue proposes
catching it *at commit time* instead, since a pre-commit nudge is cheaper to act on than a
warning discovered days later during an unrelated upstream sync.

---

## Proposed Solution

Add a lightweight local gate (pre-commit hook or `AGENTS.md`-documented manual step — pre-commit
hook preferred so it applies to human and agent commits alike) that:

1. Detects when the staged changes touch a file matched by `defaultInvariantRules`.
2. Detects whether the staged changes also touch at least one file under `test/`.
3. If (1) is true and (2) is false, blocks the commit (or prints a strong warning, if a hard
   block proves too noisy in practice) asking the author to either:
   - grep the changed symbols/strings across `test/` and update affected tests, or
   - explicitly note in the commit message why no test changes were needed (e.g.
     `no-test-touch: doc comment only`).

Keep the check narrowly scoped to the same file list `invariants.go` already tracks — this is
not a general "every commit needs a test" policy, just a guard for the specific fork-owned
surfaces that have already caused test debt twice.

This is a **process** change, not a `fork-sync verify` feature — it should live in the repo's
pre-commit tooling (see how `npm run check` is already wired into pre-commit) and be documented
in `AGENTS.md`'s "Upstream Synchronization Workflow" / fork-patch guidance, not inside the Go
CLI.

---

## Acceptance Criteria

- [ ] Pre-commit check (or documented manual step, if a hook is judged impractical) flags
      commits that touch an invariant file without touching `test/`.
- [ ] Escape hatch exists for legitimate no-test-needed changes (doc/comment-only edits) that
      doesn't require disabling the check entirely.
- [ ] Documented in `AGENTS.md`.
- [ ] Verified against the two historical offending commits (`e2f3e3041`, `171b813a2`) — the
      check should have flagged both.
