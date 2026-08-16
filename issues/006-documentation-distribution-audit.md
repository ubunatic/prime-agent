# Issue 006: Correct Fork Distribution Documentation

**Status:** Resolved

---

## Summary

The fork documentation describes GitHub Releases as the permanent and exclusive distribution architecture. The current implementation instead defaults to `https://ubunatic.com/prime-agent`, supports Cloudflare R2 when configured, and uses GitHub Releases as a fallback artifact host. The documentation must distinguish current behavior from historical release-validation notes.

## Findings

1. `packages/coding-agent/src/utils/version-check.ts` and `scripts/pack-prime-agent-release.mjs` default to `https://ubunatic.com/prime-agent`, not GitHub Releases.
2. `.github/workflows/build-binaries.yml` prefers `vars.R2_PUBLIC_BASE_URL` when present and conditionally publishes artifacts to R2 when its secrets are configured. GitHub Releases is supported without those secrets but is not the exclusive pipeline.
3. The workflow publishes stable and beta channels. The architecture guide documents only stable assets.
4. Documentation uses checkout-specific `file://` links. These fail in other clones and GitHub renderers.
5. `docs/README.md` describes issue indices as being in `docs/`, although `issues/` is a repository-root sibling.
6. The case study has duplicate `## 3` headings and includes time-bound or unverified claims, including the checked-file count and elapsed implementation time.
7. `.github/PULL_REQUEST_TEMPLATE.md` still states that pull requests require vouched contributors, despite removal of the vouch workflow and CI gate.

## Required Changes

1. Update `docs/ForkArchitecture.md` to describe the custom-domain default, optional R2 publishing, and GitHub Releases fallback accurately.
2. Document both stable and beta release assets and their channel-resolution URLs.
3. Replace all `file://` documentation links with repository-relative Markdown links.
4. Correct the documentation index description and its link to `issues/README.md`.
5. Revise the case study as a dated historical record: fix heading numbering, qualify historical measurements, and remove claims that are no longer true.
6. Either remove the vouched-contributor requirement from the PR template or restore an enforcement mechanism; the policy and implementation must agree.
7. Re-run the release and installer checks after the documentation changes, and record the validated release topology.

## Acceptance Criteria

- Documentation names `https://ubunatic.com/prime-agent` as the runtime default.
- Documentation states that R2 publication is conditional and GitHub Releases requires only the repository token.
- Stable and beta behavior are both documented.
- No `file://` links remain in `docs/` or `issues/`.
- The documented contribution policy matches active GitHub workflows and templates.

## Resolution

- Updated the architecture guide to document the custom-domain default, conditional R2 publishing, GitHub Releases fallback, and both release channels.
- Replaced checkout-specific documentation links with repository-relative Markdown links.
- Recast the GitHub Releases study as a historical record and removed stale exclusivity and measurement claims.
- Removed the obsolete vouched-contributor statement from the pull-request template.
- Verified that no `file://` links remain under `docs/` or `issues/`.
