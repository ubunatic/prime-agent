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

No code changes needed on `sync/upstream-2026-08-23` for the v0.8.0 merge.
Fork invariants are intact and automated verification is green.
