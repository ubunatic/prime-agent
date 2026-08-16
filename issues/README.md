# Issue Trackers & Status

This index tracks all specific design issues, behavioral audits, feature proposals, and bug fixes for the `ubunatic/prime-agent` fork.

| ID | Title | Status | Summary |
| :--- | :--- | :--- | :--- |
| [`001`](001-telemetry-and-call-home-audit.md) | **Telemetry & Call-Home Audit** | **Resolved** | Audited all outbound telemetry endpoints (`telemetry.ts`, `version-check.ts`). Telemetry is now disabled (`opt-in`) by default. |
| [`002`](002-non-global-installer-strategy.md) | **Non-Global Installer & Domain Setup** | **Resolved** | Updated `install.sh` to install into user space (`$HOME/.local`) without `sudo`, and configured default release domain to `ubunatic.com/prime-agent`. |
| [`003`](003-disable-forced-prime-intellect-login.md) | **Disable Forced Prime Intellect Login** | **Resolved** | Disabled forced Prime Intellect splash on onboarding; unconfigured first launch opens provider key setup directly. |
| [`004`](004-remove-npm-as-package-source.md) | **Optional NPM Publishing for Fork Releases** | **Resolved** | Added `PI_SKIP_NPM_PUBLISH=1` / `--skip-npm-publish` to `scripts/release.mjs` while preserving upstream release logic. |
| [`005`](005-fork-github-workflows-adaptation.md) | **GitHub Workflows Adaptation** | **Resolved** | Configured `build-binaries.yml` for GitHub Releases artifact publishing without R2 secrets, and removed `contribution-gate.yml`. |
| [`006`](006-documentation-distribution-audit.md) | **Fork Distribution Documentation Audit** | **Resolved** | Corrected distribution claims, portable links, and the vouch policy mismatch. |
| [`007`](007-fullscreen-wheel-scroll-without-selection.md) | **Fullscreen Wheel Scrolling Without Selection Copy** | **Resolved** | Preserved transcript wheel scrolling while disabling in-app mouse actions. |
| [`008`](008-fullscreen-click-selection.md) | **Fullscreen Word and Line Click Selection** | **Resolved** | Add double-click word selection and triple-click line selection to fullscreen in-app selection. |
