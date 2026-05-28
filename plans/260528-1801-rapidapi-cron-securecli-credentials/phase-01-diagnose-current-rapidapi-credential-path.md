---
phase: 1
title: "Diagnose Current RapidAPI Credential Path"
status: pending
priority: P1
effort: "1.5h"
dependencies: []
---

# Phase 1: Diagnose Current RapidAPI Credential Path

## Context Links

- Issue: `digitopvn/goclaw#74`
- Related closed issue: `digitopvn/goclaw#54`
- Cron payload model: `internal/store/cron_store.go`
- Cron run injection: `cmd/gateway_cron.go`
- SecureCLI lookup: `internal/tools/credentialed_exec.go`

## Overview

Prove the concrete failure mode before code. Current code already preserves cron credential user identity; #74 may be RapidAPI config, grant, env key, or legacy job payload.

## Key Insights

- `CronPayload.CredentialUserID` exists and is redacted from API responses.
- `PGCronStore.AddJob` and `SQLiteCronStore.AddJob` store explicit credential user identity.
- `makeCronJobHandler` injects that identity into cron run context.
- `lookupCredentialedBinary` resolves per-user env through `CredentialUserIDFromContext`.
- No built-in RapidAPI preset exists today.

## Requirements

- Functional: identify whether `rapidapi` is configured, granted, and populated with `RAPIDAPI_KEY`.
- Functional: identify if failing cron payload is legacy and lacks `credentialUserId`.
- Non-functional: no secret values in terminal, docs, issue comment, tests, or logs.
- Non-functional: read-only live checks unless operator explicitly approves mutation.

## Architecture

Cron creation captures credential identity in payload. Cron execution injects that identity into context. Exec tool uses context to join SecureCLI per-user env. Diagnosis must trace all three links.

```text
cron.create ctx -> payload.credentialUserId -> makeCronJobHandler ctx -> exec LookupByBinary(userID) -> RAPIDAPI_KEY env
```

## Related Code Files

- Modify: none in this phase.
- Read: `internal/store/pg/cron_crud.go`
- Read: `internal/store/sqlitestore/cron_crud.go`
- Read: `cmd/gateway_cron.go`
- Read: `internal/tools/credentialed_exec.go`
- Read: `internal/store/pg/secure_cli.go`
- Read: `internal/store/sqlitestore/secure-cli.go`

## Implementation Steps

1. Write failing characterization notes before code:
   - `rapidapi` absent from SecureCLI registry.
   - `rapidapi` present but grant missing/disabled.
   - grant present but `RAPIDAPI_KEY` missing from merged env.
   - cron payload lacks `credentialUserId`.
2. Inspect local DB or live DB only with read-only commands if credentials available:
   - SecureCLI binary rows for `rapidapi`.
   - grants for the cron agent.
   - sanitized env keys only; never reveal value.
   - cron payload JSON for target job.
3. Add or prepare unit-level characterization tests before changing product code:
   - cron handler passes `credentialUserId` through.
   - SecureCLI env merge exposes `RAPIDAPI_KEY` when configured.
   - missing user env produces actionable branch in diagnostics plan.
4. Decide root cause:
   - config-only: document operator fix and still consider preset/logging.
   - code gap: proceed Phase 2.

## Todo List

- [ ] Confirm current branch contains #54 fix.
- [ ] Confirm `rapidapi` preset absence.
- [ ] Inspect sanitized SecureCLI config path.
- [ ] Inspect failing cron payload shape if available.
- [ ] Record exact failure class.

## Success Criteria

- [ ] Root cause class documented with code/file evidence.
- [ ] No secret material displayed or persisted.
- [ ] Tests-before list finalized for Phase 2.

## Risk Assessment

- Risk: live DB access reveals secrets. Mitigation: query key names/sanitized responses only.
- Risk: legacy cron payload mistaken for current bug. Mitigation: compare job creation time against fix version.
- Risk: RapidAPI CLI command shape unknown. Mitigation: treat smoke command as operator-provided until verified.

## Security Considerations

- Never use `printenv`, `env`, or CLI debug flags to prove secret presence.
- Validate by CLI success/failure and sanitized env key metadata.

## Next Steps

- If root cause is missing preset or diagnostics, Phase 2.
- If root cause is stale cron payload only, Phase 3 documents remediation and issue reply.
