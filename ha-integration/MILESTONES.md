# Semaphore fork milestone tracker

Updated 2026-10-24. "Committed" does not mean compiled or tested.

| Milestone | Description | Status |
|---|---|---|
| M0 | Public fork and feature/postgres-ha branch | Done |
| M1.1–M1.5 | Standalone PostgreSQL coordination prototypes | External prototypes; not imported |
| M1.6 | TaskPool fail-closed dispatch and durable claim seams | Committed; not wired |
| M1.7 | Remote finalization authorization seam | Committed; not wired |
| M1.8 | Explicit runner fencing token completion entry point | Committed; token not transported |
| M1.9 | Transaction boundary for terminal fencing | Stub; always errors |
| M1.10 | Agent guidance, documentation, milestone tracker | Committed |
| M2 | Import coordinator and migrations into fork | In progress: fenced claim-lock helper committed; coordinator import pending |
| M3 | Authenticated runner assignment/report fencing token | Pending |
| M4 | Atomic fenced task and output writes | Pending |
| M5 | Heartbeat, recovery and stop routing | Pending |
| M6 | Real PostgreSQL concurrency and adversarial tests | Pending |
| M7 | HA opt-in, deployment and rollback runbook | Pending |
| M8 | AWX gap features: DAGs, approvals, RBAC | Backlog |

## Release gate
**Multi-node execution is disabled until M2–M7 are implemented, compiled and tested.** A token check before an unfenced update is insufficient.

## Current risks
- TaskPool callbacks not wired to durable PostgreSQL implementation.
- Runner protocol has no authenticated fencing-token propagation.
- Semaphore status/output writes are not transactionally fenced.
- Existing prototype schema not reconciled with current fork.
- Full fork build and PostgreSQL integration tests not executed.

## Next
Inspect runner assignment/report handlers, integrate coordinator migrations, then implement locked PostgreSQL writes and stale-runner tests. Record commit hashes and executed verification with each change.

## 2026-10-24 M2 progress
- `3f4f159`: transactional claim and node row locking against actual `ha_task_queue` schema.
- `9d27f4f`: negative-path unit tests committed, **not executed in the fork**.
- The Semaphore terminal status adapter still fails closed; runner protocol, migration import, and integration tests remain pending.
- Important: the current TaskPool callback API only returns a boolean claim, not the claim token; redesign this before HA rollout.

## 2026-10-24 typed claim progress
- `84ef5ed`: durable claim callback now returns task ID, opaque token, owner node and owner generation. Dispatch rejects malformed claims.
- `2fb5582`: fixes earlier inconsistent test callback signatures. Tests remain **unexecuted**.
- **Important safety limitation:** the claim is not yet attached to the TaskRunner or runner assignment payload, and executor handoff is not fenced. Do not enable HA mode.
- Next: carry claim through TaskRunner and remote runner protocol; implement atomic PostgreSQL completion and output writes.

## 2026-10-24 TaskRunner claim attachment
- `f0710e0`: TaskRunner retains the durable claim and exposes a defensive-copy accessor.
- `6426d97`: TaskPool attaches the exact granted claim before calling runTask.
- `ad4efbc`: adds accessor tests; not yet executed.
- No remote protocol propagation or atomic completion persistence yet. The PostgreSQL coordinator is not wired into startup; HA remains disabled.

## 2026-10-24 M3 pre-handoff gate
- `9044e35`: HA durable claim path requires an explicitly configured handoff-readiness callback before acquiring ownership.
- `11a129f`: tests for disabled-by-default gate and callback injection committed, not executed.
- This is **not** authenticated runner token transport or transactional executor fencing. A permissive callback could bypass readiness; no production enablement is authorized.
- Next: inspect actual runner assignment/report API models and wire the claim token end to end, with PostgreSQL transaction checks.

## 2026-10-24 M3 safety correction and contract
- `52b49cd`, `b240291`: hard-disable HA dispatch, regardless of readiness callback, until authenticated runner fencing and atomic status persistence exist. Legacy mode remains available.
- `b267e9a`: adds validated assignment/completion DTO contract in `ha-integration/hafence/protocol.go`; **not yet wired into HTTP runner handlers**.
- No compilation or integration test executed. Do not mark M3 complete.

## 2026-10-24 M3 runner API integration
- `9b5aa55`: optional `FencingIdentity` in runner `JobData` and `JobProgress` wire DTOs.
- `2293914`: server job assignment populates fencing identity from `TaskRunner.DurableClaim()`.
- `8b391b2`: server rejects progress reports for in-memory HA-claimed tasks pending atomic fenced persistence.
- **Not complete:** runner client does not yet echo identity; task status/output writes remain unfenced; HA dispatch remains hard disabled. Reports for HA tasks reconstructed without in-memory claims need durable DB checks before enabling HA.
- No full compile or integration tests executed.

## 2026-10-24 M3 runner echo
- `8f2cb31`: queued job retains assignment fencing identity.
- `128d1fe`: running job retains identity.
- `5514eb2`: progress report echoes identity from original assignment.
- Pending: verify protocol serialization in tests, validate runner identity and fencing claim transactionally on server, and implement atomic task/output persistence.
- HA dispatch remains hard disabled; no full compilation or integration tests executed.

## 2026-10-24 M4 transactional adapter groundwork
- `dc7f213`: strict positive decimal token parsing from runner wire contract into SQL Claim (int64).
- `0eb0f5f`: conversion and completion validation unit tests committed; not executed.
- `0f1f3a5`: CommitTerminal adapter boundary invokes caller's persistence callback under WithClaimLock; absent callback fails closed.
- **Still missing:** Semaphore-specific SQL updates for status, output, task events and durable outbox; schema verification; server handler integration; runner binding and stale-claim concurrency tests. M4 not complete; HA remains disabled.

## 2026-10-24 M4 task persistence primitive
- `636735d`: reject reports with explicit HA fencing identity on legacy runner progress path, including after in-memory task reconstruction.
- `3059eeb`: PostgreSQL task status/artifacts mutation under WithClaimLock with runner ID and nonterminal row predicates. No outbox or event write yet.
- `c1143dc`: nil transaction/database rejection tests committed, not executed.
- **Security caveat:** legacy reports without fencing identity for reconstructed HA tasks require a durable database-level HA ownership check before enabling HA. Terminal mutation helper is not production-wired; status transitions, schema and transaction semantics require integration tests.
- M4 incomplete; HA dispatch hard disabled.

## 2026-10-24 M4 durable completion outbox prototype
- `9f8873a`: isolated `ha_task_outbox` PostgreSQL schema; not registered with Semaphore migrations.
- `a491760`: `CommitTaskTerminal` atomically applies fenced task mutation and inserts idempotent completion outbox record; does not publish events or perform workflow advancement.
- `0961de8`: invalid terminal status/JSON tests committed but not executed.
- **Pending:** align schema with coordinator prototype, actual DB migration integration, event delivery worker, persisted task-event model, runner API wiring, comprehensive SQL integration tests and multi-node adversarial tests.
- M4 incomplete. HA dispatch remains disabled.

## 2026-10-24 M4 leased outbox groundwork
- `474bc2f`: isolated delivery-lease schema migration; not wired to production migrations.
- `642b025`: PostgreSQL SKIP LOCKED claim and conditional acknowledgment.
- `7b62a25`: acknowledgment fenced by delivery attempt number as well as owner and lease.
- `8b4d68b`: invalid-parameter tests committed, not executed.
- Pending: real outbox consumer/dispatch, retry/backoff and poison-event handling, authenticated API integration, PostgreSQL integration tests, migration wiring and HA rollout gates.
- Delivery is at-least-once; consumers must be idempotent. M4 incomplete; HA dispatch disabled.

## 2026-10-24 M4 outbox retries and opt-in worker
- `deb9cf6`: retry availability, dead-letter timestamp and last-error schema.
- `b2b3acb`: claim query excludes deferred and dead-letter events.
- `88d2157`: fenced failure transition with retry/dead-letter policy.
- `f5d8a21`: opt-in RunOnce delivery worker with caller-supplied idempotent handler; not registered in server lifecycle.
- **Pending:** actual Semaphore task-completed handler and transactional deduplication, production migration registration, retry backoff policy, lease extension for slow handlers, PostgreSQL integration tests, multi-node tests and HA rollout gate. No tests executed. M4 incomplete; HA disabled.

## 2026-10-24 M4 validation follow-up
- `9db9042`: prohibit output writes for non-success statuses; permit starting/stopping status transitions in fenced task mutation.
- `02ce864`: add fail-closed worker policy unit tests (not executed).
- **Remaining critical path:** compile and test with PostgreSQL; implement lease renewal for slow handlers; wire trusted, idempotent Semaphore completion handlers and actual migration lifecycle; prevent tokenless HA progress after failover using durable ownership lookup.
- HA execution remains disabled.

## 2026-10-24 M4 recovery controls
- `0e14881`: attempt-fenced outbox lease renewal primitive.
- `3034731`: dead-letter replay primitive (not API-wired; requires operator authorization and audit before exposure).
- `77f55f7`: invalid-input recovery tests committed but not executed.
- **Remaining:** wire lease renewal into worker lifecycle, implement authorized replay API, durable/idempotent Semaphore completion handler, migrations, PostgreSQL concurrency tests and end-to-end runner integration.
- M4 incomplete; HA disabled.
