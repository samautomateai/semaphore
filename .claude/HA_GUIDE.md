# HA work guidance

Read `AGENTS.md`, `.claude/CLAUDE.md`, `docs/ha-development.md` and `docs/ha-milestones.md` before modifying HA code.

- Only change the fork's `feature/postgres-ha` branch for this work.
- Keep Community compatibility; do not copy proprietary Pro code.
- PostgreSQL must be authoritative for leases and fencing.
- A pre-write authorization check is insufficient; use atomic database updates.
- Treat task stop, runner report, recovery and outbox operations as adversarial inputs.
- Never claim tests passed unless executed.
- Record milestone evidence and remaining risks.
