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
   - Remove or gate `npm run publish` step (which publishes packages to `registry.npmjs.org`).
   - Retain release tarball packaging (`release:pack`) and GitHub release tagging/uploads.
2. In GitHub Actions release workflow (`.github/workflows/build-binaries.yml`):
   - Package release tarballs and upload them to GitHub Releases and the `ubunatic.com` release mirror.
   - Do not trigger NPM registry publish steps.

---

## 3. Execution Plan

1. Update `scripts/release.mjs` to bypass NPM publishing.
2. Ensure `install.sh` exclusively operates on GitHub release tarballs and uses NPM locally for dependency installation into `$HOME/.local`.
