# Issue 004: Remove NPM as Source for `prime-agent` Package Installation

This document details the strategy for transitioning `prime-agent` installation away from publishing to or pulling the primary application binary from the public NPM registry, using NPM only to fetch runtime dependencies.

---

## 1. Goal & Rationale

- **Primary Application Distribution Source:** GitHub Releases / custom domain release mirrors (`https://ubunatic.com/prime-agent/releases`).
- **Role of NPM:** NPM will **only** be used as a package manager to fetch external runtime dependencies (e.g. `@earendil-works/pi-ai`, `get-east-asian-width`, etc.) during build/installation time.
- **Benefits:**
  - Prevents reliance on the NPM registry for hosting application binaries or release manifests.
  - Ensures complete ownership over the distribution channel under `ubunatic.com` / GitHub.

---

## 2. Technical Modifications Required

### A. Installer (`install.sh`)
1. Ensure `install.sh` downloads `prime-agent` directly as a verified tarball from GitHub Releases / custom base URL (`https://ubunatic.com/prime-agent/releases/vX.Y.Z/...`).
2. Replace any `npm install -g <package-name>` fallback invocations so the installer exclusively handles the tarball fetched from the release domain.
3. Keep `npm install --production` or `npm install --prefix ...` solely for installing declared node dependencies when unpacking into `$HOME/.local/share/prime-agent`.

### B. Release Script & CI Workflows (`scripts/release.mjs` & GitHub Actions)
1. In `scripts/release.mjs`:
   - Keep upstream `npm publish` code intact for seamless upstream syncing, but make it **optional** via a toggle flag or environment variable (e.g. `--skip-npm-publish` or `PI_SKIP_NPM_PUBLISH=1`).
   - Add first-class support for a local/fork release flow (`npm run release:pack` / release tarball output) so forks can publish releases to GitHub Releases or custom servers without requiring NPM registry publishing rights.
2. In GitHub Actions release workflow (`.github/workflows/build-binaries.yml`):
   - Support fork release builds by allowing NPM publishing to be bypassed while producing verified release tarballs and manifests for GitHub Releases and domain mirrors.

---

## 3. Execution Plan

1. Add `PI_SKIP_NPM_PUBLISH` environment variable check and `--skip-npm-publish` flag to `scripts/release.mjs`.
2. Ensure `install.sh` and release packager (`scripts/pack-prime-agent-release.mjs`) work cleanly for fork-driven releases published to GitHub Releases or `ubunatic.com/prime-agent`.

