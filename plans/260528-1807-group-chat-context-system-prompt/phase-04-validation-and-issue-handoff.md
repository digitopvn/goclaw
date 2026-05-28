---
phase: 4
title: "Validation and Issue Handoff"
status: pending
priority: P2
effort: "2h"
dependencies: [1, 2, 3]
---

# Phase 4: Validation and Issue Handoff

## Context Links

- GitHub issue: `digitopvn/goclaw#70`
- Plan path: `plans/260528-1807-group-chat-context-system-prompt/plan.md`
- Required checks: `CLAUDE.md` post-implementation checklist

## Overview

Validate the prompt-only implementation, update issue handoff, and keep docs impact explicit. Current planning task also comments issue #70 with plan summary.

## Key Insights

- Later code implementation must compile and run focused tests.
- Current user request: create worktree + branch, create plan, commit+push, reply issue with summary and filepath.

## Requirements

- Functional: focused tests pass for prompt and changed channel adapters.
- Functional: build passes after implementation.
- Functional: GitHub issue receives plan summary and path.
- Non-functional: no secret files staged; no unrelated staged files from main checkout.

## Architecture

Validation ladder:

```text
prompt tests -> channel metadata tests -> cmd consumer tests -> go build ./...
```

Issue comment includes:

- Plan path
- Scope boundaries
- Phase list
- Test gates
- Branch link

## Related Code Files

- Modify: `docs/project-changelog.md` only during implementation if code lands.
- No code file modified by planning-only commit.

## Implementation Steps

1. For current planning task: commit plan files only on `codex/issue-70-group-context-plan`.
2. Push branch.
3. Comment `digitopvn/goclaw#70` with plan summary and filepath.
4. For later implementation: run focused tests after each phase.
5. Before implementation PR: run `go build ./...`; consider `go build -tags sqliteonly ./...` because prompt code affects desktop.
6. Update changelog after implementation, not for planning-only commit unless repo convention requires it.

## Tests Before

- Not applicable to planning-only commit.
- Later implementation starts with failing prompt contract tests from Phase 1.

## Refactor

- None. This phase validates and hands off.

## Tests After

- Planning-only: `git diff --check`, `git status --short`, staged diff secret scan.
- Later implementation: `go test ./internal/agent ./cmd ./internal/channels/...`, `go build ./...`.

## Todo List

- [ ] Commit plan files only.
- [ ] Push branch.
- [ ] Comment GitHub issue #70.
- [ ] Record unresolved questions, if any.

## Success Criteria

- [ ] Branch contains only plan artifacts for issue #70.
- [ ] Issue #70 comment links `plans/260528-1807-group-chat-context-system-prompt/plan.md`.
- [ ] Plan is self-contained enough for `/ck:cook`.
- [ ] Unresolved questions section says none.

## Risk Assessment

- Risk: dirty main checkout stages unrelated files. Mitigation: use isolated worktree and explicit path staging.
- Risk: issue comment overpromises implementation. Mitigation: call it approved implementation plan, not completed feature.

## Security Considerations

- Run staged diff secret scan before commit.
- Do not include env values, tokens, or private webhook payloads in issue comment.

## Next Steps

- Recommended next command after planning: `/ck:cook /Users/duynguyen/.codex/worktrees/issue-70-group-context-plan/plans/260528-1807-group-chat-context-system-prompt/plan.md`

## Unresolved Questions

None.
