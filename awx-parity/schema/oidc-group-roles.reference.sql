-- M8.1 reference schema, NOT an applied Semaphore migration.
-- Requires migration registration, SQL dialect validation and RBAC query wiring.
-- Never write IdP-derived grants to project__user_role (manual assignments).
CREATE TABLE oidc_group_role_mapping (
    id INTEGER PRIMARY KEY,
    provider VARCHAR(128) NOT NULL,
    group_name VARCHAR(512) NOT NULL,
    project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE RESTRICT,
    UNIQUE(provider, group_name, project_id, role_id)
);
CREATE INDEX oidc_group_role_mapping_role_idx ON oidc_group_role_mapping(role_id);

CREATE TABLE oidc_user_mapped_role (
    user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    mapping_id INTEGER NOT NULL REFERENCES oidc_group_role_mapping(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, mapping_id)
);
CREATE INDEX oidc_user_mapped_role_mapping_idx ON oidc_user_mapped_role(mapping_id);

-- Synchronization must be one transaction per user/provider:
-- 1. Lock user/provider membership synchronization scope.
-- 2. Resolve only mappings matching the verified provider and group strings.
-- 3. Validate mapping project and role scope.
-- 4. Insert newly derived assignments and remove stale assignments for that provider.
-- 5. Return added/removed records for audit after commit.
-- Effective project permissions are union(manual grants, mapped grants).
-- A missing groups claim is NOT equivalent to an empty claim.
