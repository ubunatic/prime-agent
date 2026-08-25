# 013 — Proposal: Unified Project-Local (`<repo>/.prime-agent` or `<cwd>/.prime-agent`) Kickstart & Persistence for Restricted Environments

**Status**: Open (Proposal)
**Priority**: P2 (Medium)
**Severity**: Enhancement / Architecture
**Category**: Sandboxing / Portability / DX
**Related**: [[002-non-global-installer-strategy]], `dist/core/kernel/bootstrap.js`, `dist/modes/daemon/daemon-supervisor-ownership.js`

---

## 1. Problem & Motivation

In restricted environments — such as hardened rootless containers (`--read-only`, `--userns=keep-id`), CI runners with immutable user home directories, or isolated development sandboxes — Prime Agent encounters several friction points:

1. **Fragmented `$HOME`-derived paths**:
   - Configuration & auth: `~/.prime/agent/models.json` & `auth.json`
   - Session logs: `~/.prime/agent/sessions/`
   - Daemon supervisor registry: `~/.prime/supervisor-owners/`
   - Python IPython kernel venv: `$XDG_DATA_HOME/prime/agent/kernel-venv` (defaults to `~/.local/share/prime/agent/kernel-venv`)
2. **High environment variable overhead**:
   Running in a read-only or mapped user environment currently requires setting at least 5 separate environment variables (`PRIME_AGENT_CODING_AGENT_DIR`, `PRIME_AGENT_CODING_AGENT_SESSION_DIR`, `PRIME_AGENT_INTERNAL_DAEMON_SUPERVISOR_REGISTRY_DIR`, `PRIME_AGENT_KERNEL_VENV`, and `PRIME_AGENT_KERNEL_PYTHON`) plus building shell launcher wrappers.
3. **Loss of installed tools and skills on ephemeral/sandboxed runs**:
   When `/tmp` is wiped between invocations, any Python skills or packages installed during agent workflows are discarded. In project-centric workflows, tools and skills installed for a repository should naturally persist with that repository or workspace.

## 2. Proposal Specification

Introduce a unified **Project-Local Discovery & Kickstart Mode** where Prime Agent resolves state from a single cohesive directory structure, prioritizing project-local workspace roots over global `$HOME`.

### Resolution Priority Order

When resolving configuration, supervisor state, sessions, and the Python kernel venv:

1. **Explicit Environment Overrides** (`PRIME_AGENT_DIR` / `PRIME_AGENT_HOME`, or individual granular vars).
2. **Project-Local Workspace Root**:
   - Look for `.prime-agent/` or `.prime/` at `<repo-root>/.prime-agent` or `<cwd>/.prime-agent`.
3. **User XDG Base Directory**:
   - Config: `$XDG_CONFIG_HOME/prime-agent` (default `~/.config/prime-agent`)
   - Data & Kernel Venv: `$XDG_DATA_HOME/prime-agent` (default `~/.local/share/prime-agent`)
   - State & Supervisor: `$XDG_STATE_HOME/prime-agent` (default `~/.local/state/prime-agent`)
4. **Legacy Default**:
   - Fall back to `~/.prime/` and `~/.local/share/prime/agent/kernel-venv`.

### Unified Project-Local Layout (`.prime-agent/`)

When `<cwd>/.prime-agent` or `<repo>/.prime-agent` is present (or created via `prime-agent init` / `--local`):

```
<workspace-root>/.prime-agent/
├── config.json (or models.json)
├── auth.json
├── sessions/
│   └── ...
├── supervisor/
│   └── owners/
└── venv/               # persistent IPython kernel virtual environment
    ├── bin/python
    ├── pyproject.toml
    └── lib/...
```

### Benefits

1. **Restricted/Sandboxed Environments**: Mounting a single volume mount (e.g. `-v $(pwd)/.prime-agent:/work/.prime-agent:rw`) allows the entire agent state (config, supervisor, session memory, and python kernel) to function without touching `/root` or `$HOME`.
2. **Tool & Skill Persistence**: Any Python skills or local dependencies installed during an agent task remain cached and available for subsequent sessions in the same repo.
3. **Lean Kickstart with Minimal Tooling**:
   - `prime-agent` can initialize `.prime-agent/` automatically on first run.
   - If Python or `uv` is present, it seeds `.prime-agent/venv/` once.
   - If offline or pre-baked, pointing `PRIME_AGENT_DIR` to a pre-populated directory satisfies all runtime checks with a single flag.

## 3. Implementation Plan

1. **Path Resolver Refactor**:
   Consolidate path resolution in `packages/coding-agent/src/config.ts` (and kernel `bootstrap.ts`) to query `resolvePrimeAgentRoot()` first before computing sub-paths.
2. **Support Single Root Override**:
   Add `PRIME_AGENT_HOME` / `PRIME_AGENT_DIR` as top-level umbrella environment variable.
3. **Project-Local Discovery**:
   Add upward filesystem search for `.prime-agent` or `.prime` from `process.cwd()` to the nearest git/repo boundary.
4. **Verification**:
   Test in both standard local shell and hardened container environments (`--read-only` with single `.prime-agent` mount).
