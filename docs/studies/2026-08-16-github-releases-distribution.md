# Case Study: Zero-Secret Fork Releases & Distribution via GitHub Releases

**Date:** 2026-08-16  
**Scope:** Fork Infrastructure, CI/CD, Installer Adaptation, Static Analysis Tooling  
**Author:** Antigravity Pair Programming Session

---

## 1. Context & Starting State

The `ubunatic/prime-agent` repository is an independent fork of `PrimeIntellect-ai/prime-agent`. The upstream project relies on:
- Cloudflare R2 bucket credentials (`R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_ENDPOINT_URL`) for staging immutable release tarballs and manifests.
- Publishing npm packages directly to the public registry.
- `mitchellh/vouch` contributor gate actions to triage incoming PRs.

The goal of this session was to enable **Option 1 (GitHub Releases as the primary artifact host)** so that:
1. Release builds automatically attach self-contained tarballs, manifests (`latest.json`, `stable`, `SHA256SUMS`), and standalone `install.sh` scripts directly to GitHub Releases.
2. The user-local installer [`install.sh`](file:///home/uwe/git/prime-agent/install.sh) and the in-app version checker [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts) can resolve updates directly from GitHub Releases without requiring any third-party infrastructure or secrets.
3. CI runs cleanly on fork branches without vouch gating.

---

## 2. Executive Summary

- Adapted [`.github/workflows/build-binaries.yml`](file:///home/uwe/git/prime-agent/.github/workflows/build-binaries.yml) to default `PRIME_AGENT_DOWNLOAD_BASE_URL` to `https://github.com/${{ github.repository }}/releases/download` and make R2 upload steps optional when secrets are absent.
- Removed upstream contributor gate `.github/workflows/contribution-gate.yml` and eliminated vouch dependency in [`.github/workflows/ci.yml`](file:///home/uwe/git/prime-agent/.github/workflows/ci.yml).
- Updated [`scripts/pack-prime-agent-release.mjs`](file:///home/uwe/git/prime-agent/scripts/pack-prime-agent-release.mjs), [`install.sh`](file:///home/uwe/git/prime-agent/install.sh), and [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts) to handle GitHub Releases tag-based and `/latest/download` URL schemes.
- Dispatched workflow `build-binaries.yml` via GitHub Actions, published release `v0.7.3`, diagnosed live runtime errors (esbuild missing peer dep, unhandled variable declaration, double URL path prefix), and successfully verified the direct curl-to-shell installation flow.
- Expanded Biome linter configuration in [`biome.json`](file:///home/uwe/git/prime-agent/biome.json) via subagent to cover the entire `scripts/` directory, preventing future uncompiled JavaScript reference errors.

---

## 3. What Worked Well

- **GitHub CLI Integration (`gh`)**: Triggering workflow dispatches (`gh workflow run build-binaries.yml`), streaming failed run logs (`gh run view --log-failed`), and inspecting release asset payloads (`gh release view`) allowed rapid iteration without leaving the terminal.
- **Subagent Delegation for Repository-Wide Tooling**: Offloading the expansion of Biome linting across `scripts/` to a background subagent allowed cleaning 14 tooling scripts in parallel without stalling release debugging.
- **Declarative User-Local Installation**: The installer properly targets `$HOME/.local/bin`, prompts interactively, validates SHA-256 hashes against `SHA256SUMS`, and installs all four monorepo packages directly from GitHub-hosted `.tgz` archives.

---

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)

### Issue A: CLI Bundle Build Failure on `@opentelemetry/api`
- **Symptom**: The first workflow dispatch failed in `packages/coding-agent` during `npm run bundle`.
- **Cause**: `@mistralai/mistralai` imported `@opentelemetry/api`, which was an optional peer dependency not bundled in the coding agent package.
- **Detection**: Caught via `gh run view --log-failed`.
- **Fix**: Added `@opentelemetry/api` to `external: [...]` in [`packages/coding-agent/scripts/bundle.mjs`](file:///home/uwe/git/prime-agent/packages/coding-agent/scripts/bundle.mjs) ([`97f24bf9`](file:///home/uwe/git/prime-agent/packages/coding-agent/scripts/bundle.mjs)).

### Issue B: ReferenceError `manifestName is not defined`
- **Symptom**: The second workflow dispatch failed during `npm run release:pack`.
- **Cause**: When editing `scripts/pack-prime-agent-release.mjs`, `manifestName` was used in `writeJson(join(artifactsDir, manifestName), ...)` before its variable declaration.
- **Detection**: Caught by GitHub Actions run log.
- **Root Cause Analysis**: `scripts/*.mjs` were excluded from both `tsconfig.json` and `biome.json`'s `files.includes`, so `npm run check` passed locally despite containing an undeclared variable error.
- **Fix**: Re-declared `const manifestName` before usage ([`982cb06e`](file:///home/uwe/git/prime-agent/scripts/pack-prime-agent-release.mjs)), and updated [`biome.json`](file:///home/uwe/git/prime-agent/biome.json) ([`d75f096b`](file:///home/uwe/git/prime-agent/biome.json)) to lint all files in `scripts/`.

### Issue C: Double Path Prefix in Release Channel URL
- **Symptom**: `curl -fsSL .../install.sh | sh` failed with `error: could not resolve latest Prime Agent version from https://github.com/ubunatic/prime-agent/releases/download/latest/download/stable`.
- **Cause**: The base URL `https://github.com/ubunatic/prime-agent/releases/download` had `/latest/download/stable` appended to it, producing `.../releases/download/latest/download/stable` (404) instead of GitHub's actual redirection endpoint `.../releases/latest/download/stable`.
- **Detection**: Direct verification by testing the installer one-liner.
- **Fix**: Stripped the trailing `/download` suffix before appending `/latest/download/<channel>` in [`install.sh`](file:///home/uwe/git/prime-agent/install.sh) and [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts) ([`2914f07d`](file:///home/uwe/git/prime-agent/install.sh)).

---

## 5. Quality & Invariants Audit

| Invariant | Status | Notes |
| :--- | :--- | :--- |
| **Architecture & Secrets** | Passed | 100% decoupling from Cloudflare R2 secrets; release builds use default `GITHUB_TOKEN`. |
| **Static Verification** | Passed | `npm run check` passes across all 930 files (TypeScript, Biome, installer check, browser smoke). |
| **Test Coverage** | Passed | Unit tests in `packages/coding-agent/test/version-check.test.ts` verify both standard and GitHub Releases URL schemas. |
| **Idempotency** | Passed | Running `install.sh` against an existing installation cleanly replaces binaries and respects existing PATH configs. |

---

## 6. Efficiency & Velocity Assessment

- **Turnaround Time**: Initial workflow adaptation, CI trigger, diagnosis of 3 edge cases, linter expansion, and final release verification completed in ~25 minutes.
- **Live GHA Feedback Loop**: Averaged ~45s per GitHub runner build, allowing fast end-to-end iteration.

---

## 7. Key Learnings & Rules for `AGENTS.md`

1. **Tooling Scripts Must Be Linted**: Any JavaScript/MJS helper scripts in `scripts/` or `packages/*/scripts/` must be included in Biome or ESLint configs to ensure basic runtime safety (such as `noUndeclaredVariables`).
2. **Verify URL Structures against Live Hosts**: URL schemas for hosting providers (e.g. GitHub Releases redirect semantics vs. S3/R2 direct paths) must be tested with live `curl -IL` head checks before assuming path parity.

---

## 8. File & Commit Summary

### Modified / Created Files
- [`.github/workflows/build-binaries.yml`](file:///home/uwe/git/prime-agent/.github/workflows/build-binaries.yml): Fallback to GitHub Releases and optional R2 sync.
- [`.github/workflows/ci.yml`](file:///home/uwe/git/prime-agent/.github/workflows/ci.yml): Removed vouch dependency.
- `.github/workflows/contribution-gate.yml`: Deleted.
- [`install.sh`](file:///home/uwe/git/prime-agent/install.sh): Updated release channel and tarball URL resolution for GitHub.
- [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts): Manifest resolution for GitHub Releases.
- [`packages/coding-agent/test/version-check.test.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/test/version-check.test.ts): Unit tests for GitHub base URLs.
- [`packages/coding-agent/scripts/bundle.mjs`](file:///home/uwe/git/prime-agent/packages/coding-agent/scripts/bundle.mjs): Externalized `@opentelemetry/api`.
- [`scripts/pack-prime-agent-release.mjs`](file:///home/uwe/git/prime-agent/scripts/pack-prime-agent-release.mjs): Manifest tarball relative path handling.
- [`biome.json`](file:///home/uwe/git/prime-agent/biome.json): Included `scripts/` in linter rules.
- [`issues/005-fork-github-workflows-adaptation.md`](file:///home/uwe/git/prime-agent/issues/005-fork-github-workflows-adaptation.md): Documented issue resolution.
- [`issues/README.md`](file:///home/uwe/git/prime-agent/issues/README.md): Marked issue 005 as resolved.

### Git Commits
- `b9442b41` `feat(ci): support GitHub Releases artifact publishing and installer distribution (closes #5)`
- `97f24bf9` `fix(coding-agent): mark @opentelemetry/api as external in CLI bundle`
- `982cb06e` `fix(release): define manifestName in pack-prime-agent-release script`
- `d75f096b` `chore(tooling): include scripts directory in Biome linter checks`
- `2914f07d` `fix(installer): correct GitHub Releases channel resolution URL path`
