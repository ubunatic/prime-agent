# Issue 008: Fullscreen Word and Line Click Selection

**Status:** Resolved

---

## Problem

Fullscreen in-app selection supports only character-granularity dragging. It should support standard multi-click selection gestures while retaining the configured explicit-copy behavior.

## Proposed Direction

- Double-click selects the word at the clicked transcript, overlay, or dock position.
- Triple-click selects the complete visual/logical line at the clicked position.
- The resulting selection remains highlighted when `terminal.fullscreenMouseKeepSelection` is enabled.
- Copy behavior continues to follow `terminal.fullscreenMouseCopy`; explicit viewport copy also remains available.

## Acceptance Criteria

- Double-clicking selectable text selects its word without copying when automatic copy is disabled.
- Triple-clicking selectable text selects its complete line without copying when automatic copy is disabled.
- Click count resets after the multi-click interval or when the pointer position changes.
- Multi-click selection respects transcript, table, overlay, and dock selection bounds.
- Focused fullscreen tests cover word, line, and click-count reset behavior.

## Resolution

Added 500 ms same-cell multi-click tracking in fullscreen mode. Double-click selects a word and triple-click selects a line while respecting transcript, table, overlay, and dock selection bounds. The selection follows the configured persistent-selection and copy behavior. Focused fullscreen tests and `npm run check` passed.
