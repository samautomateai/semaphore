package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/require"
)

// Existing v2.20.14 schema must keep manual and mapped memberships independent.
// This test verifies the migration-backed contract before enabling Community sync.
func TestOIDCMappingSchemaPreservesManualMembership(t *testing.T) {
	store := InitConfigCreateTestStore()
	defer store.Close()

	user, err := store.CreateUserWithoutPassword(db.User{Username: "oidc-user", Name: "OIDC User", Email: "oidc@example.test"})
	require.NoError(t, err)

	project, err := store.CreateProject(db.Project{Name: "OIDC mapping test"})
	require.NoError(t, err)
	projectID := project.ID

	roleID, err := store.insert("id", "insert into role (slug, name, permissions, project_id, scope, builtin) values (?, ?, ?, ?, ?, ?)", "oidc-runner", "OIDC Runner", 1, projectID, projectID, false)
	require.NoError(t, err)

	_, err = store.exec("insert into project__user_role (project_id, user_id, role_id) values (?, ?, ?)", projectID, user.ID, roleID)
	require.NoError(t, err)

	mappingID, err := store.insert("id", "insert into group_role_mapping (provider, group_name, project_id, role_id, created) values (?, ?, ?, ?, CURRENT_TIMESTAMP)", "keycloak", "/ops", projectID, roleID)
	require.NoError(t, err)

	_, err = store.exec("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)", projectID, user.ID, mappingID)
	require.NoError(t, err)

	member, err := store.GetProjectMember(projectID, user.ID)
	require.NoError(t, err)
	require.Len(t, member.Roles, 1)
	require.Len(t, member.MappedRoles, 1)

	_, err = store.exec("delete from project__user_mapped_role where project_id=? and user_id=? and mapping_id=?", projectID, user.ID, mappingID)
	require.NoError(t, err)

	member, err = store.GetProjectMember(projectID, user.ID)
	require.NoError(t, err)
	require.Len(t, member.Roles, 1, "IdP group removal must not remove manual role")
	require.Empty(t, member.MappedRoles, "IdP group removal must remove mapped role")

	_, err = store.exec("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)", projectID, user.ID, mappingID)
	require.NoError(t, err)
	_, err = store.exec("delete from group_role_mapping where id=?", mappingID)
	require.NoError(t, err)
	member, err = store.GetProjectMember(projectID, user.ID)
	require.NoError(t, err)
	require.Len(t, member.Roles, 1)
	require.Empty(t, member.MappedRoles, "deleting a mapping must cascade its assignments")
}
