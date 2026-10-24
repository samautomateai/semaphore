# Semaphore Community HA integration

This branch is reserved for independently implemented PostgreSQL HA coordination and fencing.

Safety requirements:
- Fail closed on lease loss or database errors.
- Fence dispatch and completion by owner generation and claim token.
- Do not automatically replay dispatched non-idempotent tasks.
- Keep proprietary Pro code out of this fork.
- Treat outbox delivery as at-least-once and make handlers idempotent.

Status: standalone M1.5 Go package exists separately; it is not yet compiled into Semaphore. PostgreSQL and end-to-end integration tests are pending.

Retain the upstream MIT copyright and licence notices.