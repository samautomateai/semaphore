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
| M2 | Import coordinator and migrations into fork | Pending |
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
