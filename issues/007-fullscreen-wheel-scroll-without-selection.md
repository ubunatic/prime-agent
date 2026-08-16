# Issue 007: Fullscreen Wheel Scrolling Without Selection Copy

**Status:** Resolved

---

## Problem

`terminal.fullscreenMouse` currently controls all fullscreen mouse tracking. With it enabled, wheel events scroll the fullscreen transcript, but left-drag creates an in-app selection and copies it on release. With it disabled, the terminal owns wheel input; in alternate-screen mode this can be translated to Up/Down input and navigate editor history instead of the transcript.

The desired behavior is fullscreen transcript wheel scrolling without in-app selection or copy-on-select.

## Findings

- `Terminal.setMouseTracking()` enables `?1002` and `?1006`, so terminals report button, drag, and wheel events together.
- `TUI.handleFullscreenInput()` handles wheel events for transcript scrolling and handles left-button press, motion, and release for in-app selection and clipboard copy.
- Mouse tracking cannot generally preserve unmodified terminal-native drag selection: terminals suppress native selection while application mouse reporting is active. Shift-drag is a terminal-dependent bypass.
- Fullscreen mouse tracking does not currently back ordinary UI buttons. Left clicks are used for selection and terminal hyperlink activation.

## Proposed Direction

Add a separate `terminal.fullscreenMouseButtons` setting, defaulting to `true` for current behavior. When `fullscreenMouse` is enabled and `fullscreenMouseButtons` is `false`:

- Continue handling wheel events for fullscreen transcript scrolling.
- Ignore all left-button selection, drag, copy, and hyperlink-click behavior.

This wheel-only setting intentionally does not restore native terminal selection. It removes Prime Agent's mouse actions while retaining application wheel scrolling.

## Acceptance Criteria

- `terminal.fullscreenMouse: true` and `terminal.fullscreenMouseButtons: false` scrolls the fullscreen transcript with the wheel.
- Left drag neither creates an in-app selection nor writes to the clipboard.
- An un-dragged click on a terminal hyperlink does not open it.
- Existing default behavior remains unchanged.
- Focused TUI and coding-agent tests cover the disabled-selection behavior.

## Resolution

Added `terminal.fullscreenMouseButtons`, defaulting to `true`. Setting it to `false` retains fullscreen wheel scrolling while disabling in-app selection, clipboard copy, and hyperlink clicks. Focused TUI and settings tests cover the behavior.
