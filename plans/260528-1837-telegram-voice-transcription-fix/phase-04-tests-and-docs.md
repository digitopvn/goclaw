# Phase 04 - Tests and Docs

## Overview
- Priority: P1
- Status: Pending
- Purpose: Verify fix and sync public docs.

## Requirements
- Add regression tests for Telegram voice, STT channel override, and read_audio no-image fallback.
- Update docs to reflect actual STT provider IDs and failure behavior.
- Do not add load/stress tests.

## Related Code Files
- Modify: `docs/05-channels-messaging.md`
- Modify: `docs/03-tools-system.md` if read_audio contract changes.
- Modify: `docs/project-changelog.md`
- Run: package-level Go tests for touched packages.

## Implementation Steps
1. Run `go test ./internal/channels/telegram ./internal/audio ./internal/tools`.
2. If shared code changes, run `go test ./internal/agent ./internal/channels/... ./internal/tools`.
3. Update docs after tests pass.
4. Record issue #85 in changelog with root cause and validation commands.

## Success Criteria
- Tests pass locally.
- Docs no longer claim provider IDs that do not exist in code.
- Issue acceptance criteria are traceable to test names.

## Open Questions
- None.
