# HA fencing rollout checklist

The fork now contains TaskPool callback seams for dispatch health, durable claiming, and token-aware remote completion.

## Not production ready

- The callback interfaces are not connected to the PostgreSQL coordinator.
- No runner protocol transports or authenticates the fencing token.
- Status and output writes are not fenced in the same PostgreSQL transaction.
- A token authorization check before an unfenced write is insufficient (TOCTOU).
- TaskPool has no end-to-end HA startup wiring.
- Full Semaphore compilation and PostgreSQL integration tests have not been run.
- **Do not enable multiple active Semaphore server nodes.**

## Next implementation order

1. Extend the authenticated remote runner job assignment and completion payloads with an opaque fencing token.
2. Add a PostgreSQL transaction that locks the task lease, validates token and owner generation, and updates task status atomically.
3. Apply the same fence to output/artifact writes and terminal side effects.
4. Wire coordinator startup/heartbeat/recovery into the server lifecycle.
5. Add adversarial tests for expired leases, stale runners, duplicate reports, database failures, and concurrent node ownership.

Legacy single-node execution remains unchanged until the coordinator is explicitly configured.
