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

## Runtime adapter checkpoint — 2026-10-25

- `8c7bdd0`: added opt-in Docker/Podman CLI command builder and runner with no shell interpolation.
- `9f81468`: translates internal `if_not_present` policy to CLI `--pull=missing`.
- `5c21676`: regression coverage for restrictions, unsupported runtime, and relative workspace paths.

**Important limitations:** The adapter is not dispatched from task templates or runner APIs. Workspace canonicalization, symlink defense, trusted ownership, secrets, network policy, registry login, capability discovery, persistence, UI, cleanup, and integration tests are pending. Container flags are not a complete security boundary; Docker daemon privileges and host filesystem permissions require a separate threat-model review. Tests have not been executed. Do not enable in production.
