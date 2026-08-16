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

- [`biome.json`](../biome.json) includes root `scripts/**/*.ts`, `scripts/**/*.js`, and `scripts/**/*.mjs` so release helpers receive static checks.
- Run `npm run check` before release commands. Run the focused version-check tests when changing release URL handling.
