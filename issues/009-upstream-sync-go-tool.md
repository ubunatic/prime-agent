# Issue 009: Upstream Sync Helper (Go CLI) & Agent Workflow

**Status:** Resolved

---

## Problem

The `ubunatic/prime-agent` fork maintains several architectural invariants and localized modifications across release pipelines, installers, onboarding workflows, telemetry options, and TUI interaction handlers (documented in [`docs/ForkArchitecture.md`](../docs/ForkArchitecture.md)).

Synchronizing ongoing upstream updates from `PrimeIntellect-ai/prime-agent` manually poses risks:
1. **Accidental regression of fork invariants** (e.g. overwriting privacy defaults, non-sudo user installer behaviors, or GitHub Releases publishing logic).
2. **Merge / cherry-pick errors** across diverging package structures.
3. **Complex shell scripts** becoming brittle and hard to maintain across edge cases.

---

## Proposed Solution: Go CLI Sync Tool & Agent Workflow

Instead of complex multi-hundred-line Bash scripts, build a dedicated, compiled Go CLI helper (`tools/fork-sync` or `cmd/fork-sync`) alongside standard agent workflow instructions.

### 1. Go CLI Sync Tool Capabilities

- **Remote & Branch Validation**:
  - Automatically verifies or adds the `upstream` remote (`git@github.com:PrimeIntellect-ai/prime-agent.git` / `https://github.com/PrimeIntellect-ai/prime-agent.git`).
  - Fetches upstream references, branches, and tags.
- **Sync Plan & Diff Inspection**:
  - Inspects incoming commits between `HEAD` and `upstream/main`.
  - Audits touched paths against protected **Fork Invariant Files** (e.g., `install.sh`, `scripts/release.mjs`, `telemetry.ts`, onboarding logic, `.github/workflows/`).
- **Interactive or Automated Merge Branching**:
  - Creates an isolated sync branch (e.g. `sync/upstream-YYYY-MM-DD`).
  - Executes a clean `git merge upstream/main`.
  - In case of conflicts, classifies conflict files into "Fork Invariant" (requires fork-preserving resolution) vs. "General Upstream Feature/Fix".
- **Fork Invariant Integrity Check**:
  - Validates critical fork assertions before allowing sync commits to be finalized:
    - Telemetry opt-in default intact.
    - User-local non-sudo installer paths intact.
    - Direct provider setup without forced Prime Intellect onboarding intact.
    - GitHub Releases distribution / `PI_SKIP_NPM_PUBLISH` flag intact.

---

## Agent Instructions & Rules

Update [`AGENTS.md`](../AGENTS.md) with explicit rules for handling upstream merges:
1. Never run `git merge upstream/main` directly on `main` without a dedicated sync branch and PR verification.
2. Run the Go sync tool to inspect incoming diffs and test fork invariants.
3. Execute `npm run check` and targeted regression tests on the sync branch before merging into `main`.

---

## Acceptance Criteria

- [x] Go CLI tool implemented with commands:
  - `status` / `check`: Compare fork `main` against `upstream/main` and flag protected file overlap.
  - `start`: Fetch upstream, create `sync/upstream-<date>` branch, and initiate merge.
  - `verify`: Run invariant assertions and static checks (`npm run check`) against the working tree.
- [x] Documented upstream sync instructions in `AGENTS.md` and `issues/README.md`.
- [x] Verified Go test suite for diff analysis and invariant checking logic.

