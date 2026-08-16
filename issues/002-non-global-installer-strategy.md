# Proposal: Non-Global / Fake-Global Installation Strategy for `prime-agent`

This document analyzes the current `install.sh` workflow, which defaults to `npm install -g`, and presents detailed technical solutions for enabling non-global, user-space installations in `$HOME` (e.g. `~/.local`).

**Status:** Resolved (install.sh updated for $HOME/.local prefix & ubunatic.com domain)

---

## 1. Current State & Problem

Currently in `install.sh`:
- The installer prompts: `"Install Prime Agent vX.Y.Z globally with npm?"`
- It executes:
  ```bash
  npm install -g --no-fund --no-audit --loglevel=error --progress=false "$tarball_path"
  ```
- **Issues with Global Install:**
  1. Requires write permissions to system-wide Node directories (e.g., `/usr/local/lib/node_modules`, `/usr/lib/node_modules`), often forcing users to run `sudo npm install -g`.
  2. Pollutes system-wide npm global binaries.
  3. May fail or break in managed Node environments (nvm, fnm, system package manager) if permission defaults aren't standard.

---

## 2. Proposed Solution Architectures

### Strategy A: User-Local Prefix / "Fake Global" (`$HOME/.local` or `$XDG_DATA_HOME`)

Instead of installing to the system npm global directory, `npm` supports setting a custom prefix via the `--prefix` flag or setting `npm_config_prefix`.

#### How it works:
1. Define a user-local bin and module path:
   ```bash
   PRIME_AGENT_PREFIX="${XDG_DATA_HOME:-$HOME/.local}"
   ```
2. Run `npm install` with `--prefix`:
   ```bash
   npm install -g --prefix "$PRIME_AGENT_PREFIX" "$tarball_path"
   ```
   *Result:*
   - Binaries created at: `$PRIME_AGENT_PREFIX/bin/prime-agent`
   - Modules stored at: `$PRIME_AGENT_PREFIX/lib/node_modules/prime-agent`
   - **Zero root/sudo permissions needed!**

3. Ensure `$PRIME_AGENT_PREFIX/bin` is in the user's `$PATH` (e.g., `$HOME/.local/bin`), leveraging `install.sh`'s pre-existing shell profile update logic (`configure_standalone_node_path` / `prompt_add_standalone_node_path`).

---

### Strategy B: Direct Extraction / Local Unpack into User Space

Instead of calling `npm install -g` at all, `install.sh` can unpack the release tarball directly into a dedicated user directory (similar to how standalone Node.js is handled in `install.sh`):

1. Unpack to: `~/.local/share/prime-agent/package`
2. Run `npm install --production` inside `~/.local/share/prime-agent/package` to fetch production dependencies.
3. Symlink `~/.local/share/prime-agent/package/bin/prime-agent.js` (or wrapper) to `~/.local/bin/prime-agent`.

---

## 3. Recommended Code Modifications for `install.sh`

### Modifying `install.sh` to use User-Local Prefix by Default

In [`install.sh`](file:///home/uwe/git/prime-agent/install.sh):

1. **Update confirmation prompt & details:**
   ```bash
   # From: "Install Prime Agent v$version globally with npm?"
   # To:
   "Install Prime Agent v$version to $HOME/.local/bin?"
   ```

2. **Update `install_prime_agent_package()`:**
   ```sh
   user_prefix="${XDG_DATA_HOME:-$HOME/.local}"
   mkdir -p "$user_prefix/bin"
   
   env PRIME_AGENT_BOOTSTRAP_TOOLS_ON_INSTALL=1 \
       npm install -g --prefix "$user_prefix" \
       --no-fund --no-audit --loglevel=error --progress=false "$tarball_path"
   ```

3. **Check/Update PATH:**
   Add logic to check if `$HOME/.local/bin` is in `$PATH`, and offer automatic profile update if missing (using existing shell profile detection functions in `install.sh`).

---

## 4. Configurable Install Domain / Base URL Strategy

To allow aiming for `ubunatic.com/prime-agent` as the base download domain later, both the installer (`install.sh`) and the runtime version check logic (`version-check.ts`) need to support configurable download base URLs.

### A. Current Base URL Infrastructure
- In `install.sh`:
  `prime_agent_base_url` resolves to `${PRIME_AGENT_DOWNLOAD_BASE_URL:-__PRIME_AGENT_DOWNLOAD_BASE_URL__}`. During release build workflows, `__PRIME_AGENT_DOWNLOAD_BASE_URL__` is replaced via `scripts/pack-prime-agent-release.mjs`.
- In `version-check.ts`:
  `DEFAULT_PRIME_AGENT_DOWNLOAD_BASE_URL` defaults to `https://pub-728493de92a943e2a9b2d17b4719f318.r2.dev`, but checks `process.env.PRIME_AGENT_DOWNLOAD_BASE_URL`.

### B. Proposed Changes for Custom Domain (`ubunatic.com/prime-agent`)
1. **Default Base URL Configuration**:
   Update `DEFAULT_PRIME_AGENT_DOWNLOAD_BASE_URL` in [`packages/coding-agent/src/utils/version-check.ts`](file:///home/uwe/git/prime-agent/packages/coding-agent/src/utils/version-check.ts#L3) or make it configurable via build-time injection/settings file.
   ```ts
   const DEFAULT_PRIME_AGENT_DOWNLOAD_BASE_URL = "https://ubunatic.com/prime-agent";
   ```
2. **Release Workflow & Installer Templating**:
   Pass `--base-url "https://ubunatic.com/prime-agent"` to `scripts/pack-prime-agent-release.mjs` during the release workflow build step, causing `install.sh` to download release manifests and tarballs directly from `https://ubunatic.com/prime-agent/releases/vX.Y.Z/...`.
3. **Environment & Runtime Override**:
   Ensure `PRIME_AGENT_DOWNLOAD_BASE_URL` env var remains an explicit override for custom mirrors or local development environments.

---

## 5. Benefits
- **No `sudo` required** at any step during installation.
- Clean isolation in user `$HOME` directory.
- Fully compatible with existing node, nvm, fnm, or standalone Node setups.
- **Customizable domain deployment**: Easily brand and host installer and release artifacts under `ubunatic.com/prime-agent`.

