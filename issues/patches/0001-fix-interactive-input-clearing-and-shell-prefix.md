# Upstream Patch: Interactive Input Clearing and Shell Prefix Deletion

## Purpose

This patch contains two small independent input-editor fixes suitable for upstream contribution.

## Problems

1. `Ctrl+C` interrupted active work but retained the current draft. In an ordinary terminal, `Ctrl+C` clears the active input line as part of the interrupt.
2. The shell-mode `!` prefix is hidden from the editable prompt. After inserting `!` at the start of an existing prompt, Backspace could not delete it because the editor protected the hidden prefix, leaving the prompt stuck in shell mode.

## Changes

- Make the first `Ctrl+C` interrupt work, clear the input bar, and retain the existing second-`Ctrl+C` exit confirmation.
- Permit Backspace only when the cursor is immediately after the leading `!`; the base editor then removes the prefix and returns to the normal prompt.
- Add focused regressions for both behaviors.

## Apply

```bash
git apply issues/patches/0001-fix-interactive-input-clearing-and-shell-prefix.patch
```

## Validation

```bash
cd packages/coding-agent
npx tsx ../../node_modules/vitest/dist/cli.js --run test/interactive-mode-ctrl-c.test.ts test/custom-editor.test.ts
```
