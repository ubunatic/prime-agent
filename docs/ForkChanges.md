# Previous TypeScript fork changes

Upstream replaced the TypeScript application with the Rust workspace in
`39bc99a91`. The 45 fork-only commits below remain in Git history; their
TypeScript implementations were not copied into Rust. Each entry records the
change and whether its behavior still applies to this fork.

## User defaults and installation

- `b512cd1f9` — made npm installation user-local and corrected domain handling.
  The npm installer is gone. User-local Rust installation remains desirable,
  but the upstream Rust installer should be reviewed separately.
- `865c83c24` — disabled telemetry by default. Still needed; Rust now defaults
  telemetry off and permits an explicit settings opt-in.
- `e2f3e3041` — removed the forced Prime Intellect login from first-run setup.
  Still needed; Rust now opens provider choice first, with Prime Inference as an
  optional provider.

## Release, packaging, and CI

- `cf858525c` — made npm publishing optional for fork releases. npm publishing
  is gone; any equivalent Rust release choice is a separate policy decision.
- `de0f2c73d` — documented the npm publishing decision; historical only.
- `a5ab26530` — fixed an optional peer dependency in the TypeScript browser
  smoke check. That check is gone.
- `b9442b410` — added GitHub Releases artifact publishing and installer
  distribution. The TypeScript implementation is obsolete; Rust fork releases
  need their own release configuration if desired.
- `97f24bf91` — marked OpenTelemetry external in the TypeScript CLI bundle.
  Obsolete with the JavaScript bundle.
- `982cb06e4` — fixed the TypeScript release packer's manifest name. Obsolete
  with that packer.
- `2914f07de` — fixed the GitHub Releases channel manifest URL. The TypeScript
  URL fix is obsolete; Rust release URLs need separate review if customized.
- `88e8c0ed4` — moved TypeScript self-update to GitHub release artifacts. The
  implementation is obsolete; Rust has its own release and update system.
- `fd547843b` — removed upstream PR-gating workflows that did not fit the fork.
  Those workflows were deleted during the Rust migration; reevaluate only if
  equivalent gates return.
- `d75f096b3` — included TypeScript scripts in Biome linting. Biome is not the
  Rust source linter; no Rust change is needed.
- `388eb0f20` — documented the fork GitHub repository and release environment
  variable. The TypeScript variable is obsolete; retain only guidance relevant
  to fork-owned Rust releases.

## Interactive behavior

- `b85258780` — fixed interrupted input clearing and shell-prefix behavior in
  the TypeScript interactive agent. The old code is gone; Rust has independent
  input handling and tests.
- `171b813a2` — improved fullscreen text selection in the TypeScript TUI. The
  implementation is gone; Rust has its own terminal selection behavior.
- `a90f8a9e7` — updated TypeScript test expectations after fork changes. The
  old tests are gone.

## Fork synchronization tooling

- `7c244f6a7` — added the Go upstream-sync CLI and workflow documentation. The
  tool is gone. Auditing fork deltas remains useful, but requires Rust-aware
  tooling if automated.
- `71734ecdd` — merged upstream into the 2026-08-23 TypeScript sync branch.
  Historical merge record; the current Rust merge is recorded in Git history.
- `bdc29a8fe` — documented manual versus agent sync delegation. Process
  guidance only; replace if a Rust sync workflow is established.
- `ea49bfc12` — added direct agent invocation for sync analysis. Removed with
  the Go sync tool.
- `d5c667adb` — added a Makefile and interactive-test target for TS sync work.
  Those targets are obsolete; the Rust Makefile owns current checks.
- `de645af94` — integrated an interactive TUI smoke test into sync verification.
  Removed with the sync tool; Rust sync checks should use Rust gates.
- `5bec6dc01` — synchronized the TS model catalog. That generated catalog is
  gone; Rust owns a separate model catalog.

## Documentation, issue tracking, and release records

- `4c5a1507a` — added initial fork issue reports 001–003. Tracker history only.
- `649696b1b` — opened issue 004 about removing npm as a package source. The
  decision is reflected in the Rust migration; issue history only.
- `6a3fbb279` — opened issue 005 for GitHub workflow adaptation. Historical
  tracker record; review Rust workflows directly for fork needs.
- `10ba77d9c` — added the issue index and status badges. TS tracker metadata;
  no Rust runtime need.
- `501c3d58d` — documented the TS fork architecture. That architecture is
  obsolete; Rust-specific architecture documentation would be a replacement.
- `53f6bdeb7` — documented installed TS fork behavior in a case study.
  Historical implementation detail only.
- `c00104deb` — documented TS GitHub Releases distribution in a case study.
  The rationale may inform Rust releases, but the implementation is obsolete.
- `a8a0f397d` — corrected TS distribution documentation. Historical only.
- `d1d221922` — explained prompt context and the continual harness. The
  concepts survive, but TS-specific documentation needs replacement to describe
  Rust accurately.
- `347a01494` — recorded a v0.8.0 TypeScript sync test-debt pitfall. Historical
  TS context, not a Rust code requirement.
- `0c0f1d6b3` — replaced an inaccurate v0.8.0 fork-impact report with a
  verified TS review. Historical only.
- `d748e1448` — documented review scoping for large upstream syncs. The advice
  remains useful for Rust syncs.
- `a84ef3e1d` — opened test-debt issues and reopened Go sync-tool issue 009.
  Tracker history; the Go tool is gone.
- `223babce8` — filed issue 013 proposing a unified project-local kickstart.
  Product planning record only.
- `c0c1310ff` — merged the 2026-08-23 TS sync into the fork main branch.
  Historical merge record only.

## Changelog bookkeeping

- `9c3b9357e` — added a TS `[Unreleased]` section; release bookkeeping only.
- `045bb2fc3` — prepared TS release v0.7.5; historical release record.
- `f925d07e8` — added another TS `[Unreleased]` section; bookkeeping only.
- `1cd142c52` — prepared TS release v0.7.4; historical release record.
- `cea733206` — added a TS `[Unreleased]` section; bookkeeping only.
- `b57666f68` — prepared TS release v0.7.3; historical release record.

The earlier TypeScript fork history remains reachable through the merge
commit's first parent. Only the telemetry default and provider-first onboarding
behavior were carried into Rust for this update.
