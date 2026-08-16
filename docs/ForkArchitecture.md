# Fork Architecture & Distribution Guide

This document captures the permanent architecture, distribution design, and design decisions for the `ubunatic/prime-agent` fork.

---

## 1. Core Architectural Pillars

The fork modifies the upstream `PrimeIntellect-ai/prime-agent` codebase with four core invariants:

```
+-----------------------------------------------------------------------------------+
|                            ubunatic/prime-agent                                   |
+--------------------------+--------------------------------+-----------------------+
| 1. Privacy First         | 2. User-Local Distribution     | 3. Direct Onboarding  |
| - Telemetry: disabled    | - Target: $HOME/.local/bin     | - No forced logins    |
|   by default (opt-in)    | - No sudo / root required      | - Standard provider   |
| - Traces: private        | - Standalone tarball packaging |   API key selection   |
+--------------------------+--------------------------------+-----------------------+
|                                4. Zero-Secret CI/CD                               |
|   - GitHub Actions publishes release tarballs directly to GitHub Releases         |
|   - Automatic resolution of 'latest' / 'stable' via GitHub Releases API / redirect|
|   - Optional npm publishing (PI_SKIP_NPM_PUBLISH=1)                               |
+-----------------------------------------------------------------------------------+
```

---

## 2. Release & Distribution Pipeline

### Asset Layout in GitHub Releases
When a release tag (e.g. `v0.7.3`) is created or workflow `build-binaries.yml` is dispatched, the following assets are attached to the GitHub Release:

| File | Purpose |
| :--- | :--- |
| `prime-agent-<ver>.tgz` | Full coding-agent bundle including runtime packages. |
| `prime-agent-ai-<ver>.tgz` | Core AI provider and model integration package. |
| `prime-agent-core-<ver>.tgz` | Agent execution kernel and session management. |
| `prime-agent-tui-<ver>.tgz` | Terminal UI renderer and input components. |
| `SHA256SUMS` | Cryptographic SHA-256 manifest of all packaged tarballs. |
| `latest.json` | Version check manifest for in-app updates. |
| `stable` | Plain text version string (e.g., `v0.7.3`) for channel discovery. |
| `install.sh` | Standalone installer pre-configured for that release tag. |

### URL Resolution Architecture
- **Tarball Base URL**: `https://github.com/ubunatic/prime-agent/releases/download/v<version>/<tarball>`
- **Channel Version Endpoint**: `https://github.com/ubunatic/prime-agent/releases/latest/download/stable`
- **In-App Update Manifest**: `https://github.com/ubunatic/prime-agent/releases/latest/download/latest.json`

---

## 3. Tooling & Static Analysis Invariants

To avoid runtime reference errors in release packaging:
- **Biome Linter Configuration**: [`biome.json`](file:///home/uwe/git/prime-agent/biome.json) must explicitly include all scripts in `scripts/` (`scripts/**/*.ts`, `scripts/**/*.js`, `scripts/**/*.mjs`).
- **Release Verification**: Always run `npm run check` prior to running `npm run release:patch` or `npm run release:minor`.
