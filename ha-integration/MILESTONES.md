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
