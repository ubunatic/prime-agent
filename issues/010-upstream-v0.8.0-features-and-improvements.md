# Upstream v0.8.0 Merge — Fork Impact Review

**Status:** Resolved — no fork adjustments required.

The auto-generated report originally committed here was not real analysis: the
pre-merge agent ran in a read-only sandbox, failed to write its report, and its
raw stdout (an error message plus a terse bullet list) was committed verbatim
by the fallback path in `tools/fork-sync/agent.go`. This replaces it with a
verified review.

## Verification performed

- `tools/fork-sync verify` — all 5 fork invariants (Telemetry & Privacy,
  User-Local Installer, Direct Provider Onboarding, Release & Publishing,
  GitHub Workflows) **PASS** against the merged tree.
- Confirmed `packages/coding-agent/src/core/telemetry.ts` still honors
  `PI_OFFLINE`, `DO_NOT_TRACK`, `PRIME_AGENT_TELEMETRY`, opt-in by default.
- Confirmed `scripts/release.mjs` still supports `PI_SKIP_NPM_PUBLISH` and
  `--skip-npm-publish`.
- Confirmed `packages/tui/src/mouse.ts` / `fullscreen.ts` (TUI mouse
  scroll/selection) still match the invariant pattern; upstream touched
  `packages/tui/src/mouse.ts` (+14/-14) in this merge but the invariant check
  still passes.
- `go build ./... && go test ./...` in `tools/fork-sync` — clean.

## Risk areas flagged by the broken report

- **Daemon schema revision 16→22** (`daemon-protocol.ts`,
  `DAEMON_SCHEMA_REVISION = 22`) and **MCP credential/config migrations**
  (`daemon-mode.ts`, `daemon-supervisor.ts`, `mcp-manager.ts`, etc.): checked
  fork commit history against these paths — the fork has never carried
  patches touching daemon protocol/schema or MCP internals. These are
  entirely upstream-owned code paths with no fork invariant coverage needed.
  No action required.
- **New changelog-fragment release flow**
  (`.github/workflows/changelog-fragment.yml`,
  `scripts/lib/changelog-fragments.mjs`): the fork already removed the
  upstream PR-gating workflow for this in `fd547843b` ("chore(ci): remove
  upstream PR gating workflows"). No further action.

## Conclusion

No code changes needed on `sync/upstream-2026-08-23` for the v0.8.0 merge
itself. Fork invariants are intact.

## Follow-up: pre-existing CI failures surfaced during verification

Pushing the branch surfaced 19 `coding-agent` test failures in CI. Investigation
(reproducing each failure at the pre-merge tip `7c244f6a7`) showed **none were
caused by the v0.8.0 merge** — all 19 already failed before the merge. They were
test debt from two earlier fork commits that changed production behavior
without updating every test that asserted the old behavior:

- `e2f3e3041` ("disable forced prime intellect login", closes #3) changed
  `LOGIN_RECOVERY_MESSAGE`/`getProviderLoginHelp()` text, removed the auto
  Prime Inference login fallback in `runOnboardingFlow`, and removed the
  Prime Inference default-model preference in `model-resolver.ts` — but only
  updated 2 of the ~9 affected test files.
- `171b813a2` ("improve fullscreen text selection", fixes #8) added
  `mouseButtons`/`mouseCopy`/`mouseKeepSelection` to `applyFullscreen`'s
  `settingsManager` calls but missed one regression test's mock.

Fixed in `a90f8a9e7`: stale assertions were updated to match intended fork
behavior, and tests exercising now-deleted code paths (forced Prime CLI
splash, auto Prime Inference login) were rewritten to positively assert those
routes are *not* taken, rather than removed — see
[`docs/ForkArchitecture.md`](../docs/ForkArchitecture.md#pitfall-test-debt-from-behavior-changing-fork-patches)
for the general pattern and how to avoid it.

**Status:** Resolved. CI green on `a90f8a9e7`.
