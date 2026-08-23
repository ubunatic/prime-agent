# Issue 011: Detect Test Debt from Behavior-Changing Fork Patches in `fork-sync verify`

**Status:** Open

---

## Problem

`tools/fork-sync verify` (see [`invariants.go`](../tools/fork-sync/invariants.go)) only
asserts that fork-owned files still **exist and match an expected pattern**. It does not
assert that every test which asserted the *old* upstream behavior was updated when a fork
commit intentionally changed that behavior.

This surfaced concretely during the v0.8.0 upstream sync
([`issues/010`](010-upstream-v0.8.0-features-and-improvements.md)): two earlier fork commits —
`e2f3e3041` ("disable forced prime intellect login", closes #3) and `171b813a2` ("improve
fullscreen text selection", fixes #8) — each changed production behavior in an
invariant-covered file but only updated a subset of the tests that asserted the removed/changed
behavior. `fork-sync verify` reported all invariants green both before and after those commits;
the gap was only caught 7 days later, by an unrelated upstream sync running the full `vitest`
suite in CI, as 19 failures that looked like merge fallout. See
[`docs/ForkArchitecture.md#pitfall-test-debt-from-behavior-changing-fork-patches`](../docs/ForkArchitecture.md#pitfall-test-debt-from-behavior-changing-fork-patches)
for the full writeup of the diagnostic method used to separate fork test debt from real merge
regressions.

---

## Proposed Solution

Extend `tools/fork-sync verify` (or add a new `fork-sync audit-tests` subcommand) with a
best-effort heuristic:

1. For each file matched by `defaultInvariantRules` in `invariants.go`, find its most recent
   commit that changed an exported symbol or string literal (`git log -1 -S<symbol>` per
   exported identifier, or a simpler whole-file "last commit touching this file" as a first
   pass).
2. Check whether that same commit (or a commit within some small window) also touched at least
   one file under `test/` referencing that file's module path.
3. If not, emit a **warning** (not a hard failure — false positives are likely) listing the
   invariant file, the commit, and a reminder to grep `test/` for the changed symbols/strings
   before considering the patch complete.

This does not need to be exhaustive or precise — it only needs to catch the "changed behavior,
touched zero test files" case that both root-cause commits in `010` exhibited, and it should
run fast enough to be part of `verify --run-checks` without materially slowing it down.

---

## Acceptance Criteria

- [ ] `fork-sync verify` (or a new subcommand) warns when an invariant-covered file changes
      without any `test/` file changing in the same commit.
- [ ] Warning output includes the commit hash, the invariant file, and a pointer to
      `docs/ForkArchitecture.md`'s test-debt pitfall section.
- [ ] False-positive rate is low enough that the warning stays actionable (e.g. doc-only or
      comment-only changes to an invariant file should not warn).
- [ ] Documented in `AGENTS.md`'s Sync Tool Usage section.
