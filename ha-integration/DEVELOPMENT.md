# HA development guide

Goal: independently implement PostgreSQL-coordinated high availability in Semaphore Community without proprietary Pro code.

## Invariants
- Fail closed on coordinator/lease failure.
- A claim includes task ID, fencing token, owner node and owner generation.
- The authenticated runner must echo its issued token; never infer the current token.
- Verify ownership and update task status/output in one locked PostgreSQL transaction.
- Never automatically replay dispatched non-idempotent work.
- Outbox is at-least-once; handlers must be idempotent.
- Keep the original MIT notices and preserve upstream compatibility.

## Workflow
Read root AGENTS.md, .claude/CLAUDE.md, .claude/HA_GUIDE.md, and ha-integration/MILESTONES.md. Work on feature/postgres-ha. Make focused changes, add tests, run go test ./... and go vet ./... where possible, and distinguish committed from verified. Do not enable multi-node execution before the release gate.

## Code map
- services/tasks/TaskPool.go: dispatch and completion seams
- ha-integration/hafence/terminal.go: incomplete fence transaction stub
- ha-integration/FENCING.md: safety gate
- ha-integration/MILESTONES.md: progress and blockers
