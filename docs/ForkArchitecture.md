# Fork Architecture & Distribution Guide

This document records the current distribution architecture and operating invariants for the `ubunatic/prime-agent` fork.

---

## Core Architectural Pillars

1. **Privacy first.** Product telemetry and trace sharing are opt-in.
2. **User-local installation.** The installer targets `$HOME/.local/bin` (or `$XDG_DATA_HOME/bin`) and does not require a global npm installation.
3. **Direct provider onboarding.** Initial setup supports standard provider API keys without requiring a Prime Intellect account.
4. **Flexible, secret-free-capable distribution.** The default runtime release base is GitHub Releases. Release builds always publish GitHub Release assets with `GITHUB_TOKEN`; R2 publication is optional and occurs only when R2 configuration is supplied.

---

## Release & Distribution Pipeline

### Release bases

- **Runtime default:** `https://github.com/ubunatic/prime-agent/releases/download`
- **Optional custom mirror:** `PRIME_AGENT_DOWNLOAD_BASE_URL`
- **Optional release mirror:** `vars.R2_PUBLIC_BASE_URL`, with R2 uploads enabled only when the necessary secrets are configured.

The installer and version checker accept `PRIME_AGENT_DOWNLOAD_BASE_URL` to select a mirror. For custom-domain and R2 bases, stable metadata is served from `latest.json`, while artifacts are served below `releases/v<version>/`. For GitHub Releases, stable metadata is served through GitHub's `releases/latest/download/` redirect and production artifacts are served below `releases/download/v<version>/`.

### Release channels and assets

| Channel | GitHub Release tag | Metadata | Assets |
| :--- | :--- | :--- | :--- |
| Stable | `v<version>` | `stable`, `latest.json` | `prime-agent-<version>.tgz`, `prime-agent-ai-<version>.tgz`, `prime-agent-core-<version>.tgz`, `prime-agent-tui-<version>.tgz`, `SHA256SUMS`, `install.sh`, `install-beta.sh` |
| Beta | `beta` | `beta`, `beta.json` | The same tarball, checksum, and installer asset set for the current beta build |

The stable GitHub endpoints are:

- Channel: `https://github.com/<owner>/<repo>/releases/latest/download/stable`
- Manifest: `https://github.com/<owner>/<repo>/releases/latest/download/latest.json`
- Tarball: `https://github.com/<owner>/<repo>/releases/download/v<version>/prime-agent-<version>.tgz`

The beta channel metadata is served from `https://github.com/<owner>/<repo>/releases/beta/download/beta` and `beta.json`.

---

## Tooling and Verification

- [`tools/fork-sync`](../tools/fork-sync): Standalone Go CLI tool managing upstream merge workflows (`status`, `start`, `verify`) and invariant integrity testing.
- [`biome.json`](../biome.json) includes root `scripts/**/*.ts`, `scripts/**/*.js`, and `scripts/**/*.mjs` so release helpers receive static checks.
- Run `npm run check` before release commands. Run the focused version-check tests when changing release URL handling.
- See [`AGENTS.md`](../AGENTS.md#reviewing-a-sync-branch) for how to scope `/ultrareview` on a sync
  branch — a full upstream merge diff routinely exceeds its size limits.

---

## Pitfall: Test Debt from Behavior-Changing Fork Patches

`tools/fork-sync verify` only asserts that fork-owned *files* still exist and
match their expected pattern (see `tools/fork-sync/invariants.go`). It does
**not** assert that every test asserting the *old* upstream behavior was
updated when a fork commit intentionally changed that behavior.

This bit twice in the same session (documented in
[`issues/010`](../issues/010-upstream-v0.8.0-features-and-improvements.md)):
commit `e2f3e3041` ("disable forced prime intellect login", closes #3) changed
login-error text, deleted the auto Prime Inference login fallback, and
removed a default-model preference — but only updated 2 of the ~9 test files
that asserted the removed behavior. Commit `171b813a2` ("improve fullscreen
text selection", fixes #8) added new `settingsManager` calls but missed one
test's mock. Both went unnoticed until an unrelated upstream sync ran the
full test suite in CI and surfaced them as failures — 19 failures that looked
like merge fallout but were pre-existing.

**When making a fork patch that changes existing behavior** (not just adding
new files covered by an invariant rule):

1. Grep the string literals / function names you're changing across `test/`
   before committing — don't rely on `npm run check` or `fork-sync verify`
   catching it; those don't run the full `vitest` suite.
2. If you intentionally remove a code path, either delete the tests that only
   exercised it, or rewrite them to assert the path is *not* taken (this
   catches future regressions where the old path silently comes back).
3. When diagnosing a test failure after an upstream sync, reproduce it at the
   sync branch's pre-merge tip first. If it already fails there, it's fork
   test debt, not a merge regression — fix it directly rather than filing an
   upstream bug ticket.

