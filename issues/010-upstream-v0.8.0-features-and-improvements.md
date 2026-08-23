Unable to write the report because the workspace is mounted read-only and approval escalation is disabled. No files were changed.

Target: `/home/uwe/git/prime-agent/issues/010-upstream-v0.8.0-features-and-improvements.md`

The completed analysis found:

- Telemetry remains opt-in.
- TUI mouse scrolling and selection behavior remains preserved.
- `PI_SKIP_NPM_PUBLISH` and `--skip-npm-publish` remain supported.
- Installer and direct-provider onboarding invariants remain intact.
- Main risks are daemon schema revision 16→22, MCP credential/configuration migrations, and the new changelog-fragment release flow.