# M9 — Execution Environment Runtime

Status: **IN PROGRESS — definition hardening only; runtime not yet wired**.

## Objective
Provide digest-pinned, isolated Linux OCI execution environments for Ansible jobs, with project-scoped definitions and explicit runner capabilities. Windows native execution remains separate. HA parked; Keycloak OIDC only.

## Current changes
- `6206127`: EE definition includes project ID and optional registry credential ID; image validation rejects query/fragment, absolute path, backslash and double-slash forms.
- `bb27243`: regression tests for malicious image references.
- Earlier `db/ExecutionEnvironment.go` validates digest pinning and pull policy.

## Functional deliverables remaining
1. Schema migration and SQL CRUD with project-scoped access and authorization.
2. API for CRUD, list, selection and deletion-reference checks; UI for management and template selection.
3. Runner capability negotiation; reject unsupported container jobs.
4. Podman/Docker adapter with explicit argument vectors (never shell strings), digest verification and restricted runtime identity.
5. Per-job filesystem and network isolation; secure credential mounts, no privileged mode, resource limits and guaranteed cleanup.
6. Registry authentication without exposing tokens in logs or process arguments.
7. Cancellation, timeout, exit-code propagation, streaming logs and audit metadata.
8. Unit, integration, security and end-to-end validation on Go 1.26.4.

**Not complete or compiled:** these commits do not claim working execution environments.
