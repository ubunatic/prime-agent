# Proposal: Adapt GitHub Workflows for Fork-Friendly CI & Release Publishing

This document analyzes the existing GitHub Actions workflows in `.github/workflows/` and proposes updates to ensure clean CI execution and release publishing on the `ubunatic/prime-agent` fork without requiring third-party secrets (such as Cloudflare R2 credentials).

---

## 1. Audit of Workflows

| Workflow | Path | Status & Fork Compatibility | Recommended Action |
| :--- | :--- | :--- | :--- |
| **CI** | `.github/workflows/ci.yml` | **Compatible**<br>Runs linting (`npm run check`) and unit test shards. Works cleanly on GitHub runners without extra secrets. | Keep enabled. |
| **Release Prime Agent** | `.github/workflows/build-binaries.yml` | **Incompatible (Missing Secrets)**<br>Triggers on release tags (`v*`). Attemps to sync release tarballs to a Cloudflare R2 bucket using `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_BUCKET`. Will fail without these secrets. | Update to fallback to GitHub Releases / custom domain release host (`ubunatic.com/prime-agent`). |
| **Contributor Gate** | `.github/workflows/contribution-gate.yml` | **Unnecessary**<br>Uses `mitchellh/vouch` to check contributor trust for PRs on the main repository. | Remove or disable for the fork. |
| **Nightly Process Stress** | `.github/workflows/nightly-process-stress.yml` | **Compatible**<br>Runs process stress tests on a schedule. | Keep enabled or run manually. |

---

## 2. Technical Modifications Strategy

### A. Fallback Release Publishing in `build-binaries.yml`
Update the `publish` job in [`build-binaries.yml`](file:///home/uwe/git/prime-agent/.github/workflows/build-binaries.yml) so that if R2 bucket credentials are absent:
1. It creates/updates the GitHub Release directly using `ncipollo/release-action` or `gh release create`.
2. Attach all packaged release tarballs (`prime-agent-vX.Y.Z.tgz`) and release manifests (`latest.json`, `beta.json`, `SHA256SUMS`) as assets to the GitHub Release.

### B. Disable `contribution-gate.yml`
Remove or comment out `contribution-gate.yml` to prevent PR gating overhead.

---

## 3. Execution Plan

1. Create issue file `issues/005-fork-github-workflows-adaptation.md`.
2. Update `build-binaries.yml` to support GitHub Releases attachment when R2 secrets are not configured.
3. Remove or disable `contribution-gate.yml`.
