# AWX parity M8.1–M8.6 — compilation and validation gate

Status: **NOT VERIFIED / NOT ACCEPTED** (2026-10-25).

Scope: M8.1 Keycloak OIDC/RBAC, M8.2 runner dispatch, M8.3 scheduling, M8.4 notifications/audit, M8.5 project backup/restore, M8.6 inventory/credentials. HA excluded; SAML excluded.

## Source fixes in this pass
- `6e9bba6`: inventory SQL UPDATE now requires matching project ID.
- `56d5f73`: path regression tests cover valid dot-prefixed names and parent traversal.
- `4992604`: CI workflow includes backup regression tests, full Go compilation, and full Go tests.

## Validation to run in GitHub Actions
1. `go test -count=1 -timeout=10m ./db/sql ./pro/api ./pro/services/server ./pro/db/factory`
2. `go test -count=1 -timeout=10m ./db ./services/tasks ./services/runners`
3. `go test -count=1 -timeout=10m ./services/schedules ./api/projects`
4. `go test -count=1 -timeout=10m ./services/audit ./services/project ./services/export`
5. `go test -run '^$' ./...` (all-package compilation)
6. `go test -count=1 -timeout=20m ./...` (full suite)

The current local environment has Go 1.23.2, below the repository's Go 1.26.4 requirement. The GitHub workflow has been committed but no successful execution is established. A GitHub workflow-run lookup on an earlier branch commit returned no PR-triggered runs; this does not prove no push-triggered run exists.

## Open acceptance blockers
- M8.1: live Keycloak sign-in/group mapping, role revocation and RBAC tests.
- M8.2: runner fleet failover/capacity integration and Windows execution validation.
- M8.3: missed runs, concurrency, DST integration and audit/task-parameter validation.
- M8.4: outbound egress/SSRF restrictions, delivery retry and live provider checks.
- M8.5: atomic restore/recovery, complete references and secret-handling verification.
- M8.6: credential cross-project ownership, symlink containment and execution/rotation tests.

Do not mark milestones complete or merge until CI is green and security/operator acceptance is recorded.
