package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/require"
)

func TestOIDCGroupMappingSyncRevocation(t *testing.T) {
	store := InitConfigCreateTestStore()
	defer store.Close()
	user, err := store.CreateUserWithoutPassword(db.User{Username:"mapped-user", Name:"Mapped User", Email:"mapped@example.test"})
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name:"Mapped Roles"})
	require.NoError(t, err)
	roleID, err := store.insert("id", "insert into role (slug,name,permissions,project_id,scope,builtin) values (?,?,?,?,?,?)", "ops-runner", "Ops Runner", 1, project.ID, project.ID, false)
	require.NoError(t, err)
	_, err = store.exec("insert into project__user_role (project_id,user_id,role_id) values (?,?,?)", project.ID, user.ID, roleID)
	require.NoError(t, err)
	mappings := OIDCGroupMappingStore{DB:store}
	mapping, err := mappings.CreateRoleMapping(db.RoleMapping{Provider:"keycloak", Group:"/ops", ProjectID:project.ID, RoleID:roleID})
	require.NoError(t, err)
	require.Positive(t, mapping.ID)
	added, removed, err := mappings.SyncMappedRoles(user.ID, "keycloak", []string{"/ops"})
	require.NoError(t, err)
	require.Len(t, added, 1)
	require.Empty(t, removed)
	added, removed, err = mappings.SyncMappedRoles(user.ID, "keycloak", []string{"/ops"})
	require.NoError(t, err)
	require.Empty(t, added)
	require.Empty(t, removed)
	added, removed, err = mappings.SyncMappedRoles(user.ID, "keycloak", []string{})
	require.NoError(t, err)
	require.Empty(t, added)
	require.Len(t, removed, 1)
	member, err := store.GetProjectMember(project.ID, user.ID)
	require.NoError(t, err)
	require.Len(t, member.Roles, 1, "manual assignment survives OIDC revocation")
	require.Empty(t, member.MappedRoles)
}
