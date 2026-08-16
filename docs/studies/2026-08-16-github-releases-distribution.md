# Case Study: Adding GitHub Releases Distribution

**Date:** 2026-08-16  
**Scope:** Fork infrastructure, CI/CD, installer adaptation, and static-analysis tooling

---

## Context

This is a historical record of work that added GitHub Releases as a release-artifact host for the `ubunatic/prime-agent` fork. It does not define the current distribution architecture; see the [Fork Architecture & Distribution Guide](../ForkArchitecture.md) for that.

At the time, the work aimed to make release assets available without R2 credentials by attaching tarballs, manifests, checksums, and rendered installers to GitHub Releases. It also removed the vouch workflow from CI.

## Implemented Changes

- Updated [`build-binaries.yml`](../../.github/workflows/build-binaries.yml) to use a GitHub Releases download base when `R2_PUBLIC_BASE_URL` is not configured, while retaining conditional R2 publication.
- Updated [`install.sh`](../../install.sh), [`version-check.ts`](../../packages/coding-agent/src/utils/version-check.ts), and [`pack-prime-agent-release.mjs`](../../scripts/pack-prime-agent-release.mjs) to support GitHub Releases URL shapes.
- Removed `.github/workflows/contribution-gate.yml` and the vouch dependency from [`ci.yml`](../../.github/workflows/ci.yml).
- Added the [`@opentelemetry/api`](../../packages/coding-agent/scripts/bundle.mjs) bundle external, corrected the packer's `manifestName` declaration, and expanded [`biome.json`](../../biome.json) to include root release scripts.

## Release Behavior

GitHub Releases can publish stable and beta artifact sets using the repository `GITHUB_TOKEN`. R2 publication remains available when R2 configuration is supplied. After live validation found the planned custom-domain manifest unavailable, GitHub Releases became the runtime default; custom mirrors remain supported through configuration.

The installer validates downloaded tarballs against `SHA256SUMS` and installs into user-local npm paths. The version-check tests cover both custom-base and GitHub Releases manifest URL schemes.

## Historical Failures and Resolutions

1. **Bundle failure on `@opentelemetry/api`.** The Mistral SDK's optional peer dependency was not bundled; it was marked external in [`bundle.mjs`](../../packages/coding-agent/scripts/bundle.mjs).
2. **`manifestName` reference error.** The packer used the variable before declaration; the declaration was moved before use and root scripts were added to Biome coverage.
3. **Malformed GitHub channel URL.** The installer and version checker initially constructed `releases/download/latest/download/...`; they now strip the `download` suffix before constructing GitHub channel URLs.

## Lessons

1. Include release helper scripts in static analysis.
2. Verify hosting-provider URL semantics with live redirects or focused tests.
3. Keep dated implementation notes separate from current architecture documentation.
