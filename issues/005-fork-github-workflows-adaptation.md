# Proposal: Adapt GitHub Workflows for Fork-Friendly CI & Release Publishing

This document analyzes the existing GitHub Actions workflows in `.github/workflows/` and proposes updates to ensure clean CI execution and release publishing on the `ubunatic/prime-agent` fork without requiring third-party secrets (such as Cloudflare R2 credentials).

**Status:** Resolved

---

## 1. Audit of Workflows

| Workflow | Path | Status & Fork Compatibility | Recommended Action |
| :--- | :--- | :--- | :--- |
| **CI** | `.github/workflows/ci.yml` | **Compatible**<br>Removed upstream `mitchellh/vouch` contributor gate check so fork CI runs smoothly on all PRs/branches. | Updated to remove vouch gating. |
| **Release Prime Agent** | `.github/workflows/build-binaries.yml` | **Resolved**<br>Now defaults `PRIME_AGENT_DOWNLOAD_BASE_URL` to `https://github.com/${{ github.repository }}/releases/download` when `R2_PUBLIC_BASE_URL` is absent, makes R2 syncing optional if credentials are missing, and attaches all packaged release tarballs and installers directly to the GitHub Release. | Updated for GitHub Releases asset publishing. |
| **Contributor Gate** | `.github/workflows/contribution-gate.yml` | **Removed**<br>Removed upstream vouch contribution gating workflow. | Removed. |
| **Nightly Process Stress** | `.github/workflows/nightly-process-stress.yml` | **Compatible**<br>Runs process stress tests on a schedule. | Keep enabled or run manually. |

---

## 2. Technical Modifications Strategy

### A. Fallback Release Publishing in `build-binaries.yml`
Updated [`build-binaries.yml`](file:///home/uwe/git/prime-agent/.github/workflows/build-binaries.yml):
1. Sets `PRIME_AGENT_DOWNLOAD_BASE_URL` to `https://github.com/${{ github.repository }}/releases/download` by default.
2. Guards R2 upload steps with `env.R2_BUCKET != '' && env.R2_ENDPOINT_URL != ''`.
3. Attaches packaged release tarballs, manifests (`latest.json`, `beta.json`, `SHA256SUMS`), and rendered `install.sh` scripts directly to the GitHub Release via `gh release create` / `gh release upload`.

### B. Disable `contribution-gate.yml` & `ci.yml` Vouch Gate
Removed `contribution-gate.yml` and eliminated the vouch dependency in `ci.yml`.

---

## 3. Implementation Verification
- [`install.sh`](file:///home/uwe/git/prime-agent/install.sh), [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts), and [`scripts/pack-prime-agent-release.mjs`](file:///home/uwe/git/prime-agent/scripts/pack-prime-agent-release.mjs) aligned to work seamlessly with GitHub Releases tag-based asset download paths.
- Unit tests verified passing in `packages/coding-agent/test/version-check.test.ts`.
