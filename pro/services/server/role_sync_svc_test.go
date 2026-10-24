package server

import (
	"errors"
	"reflect"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

type recordingMappingStore struct {
	calls int
	userID int
	provider string
	groups []string
	err error
}

func (s *recordingMappingStore) GetRoleMappings() ([]db.RoleMapping, error) { return nil, nil }
func (s *recordingMappingStore) CreateRoleMapping(m db.RoleMapping) (db.RoleMapping, error) { return m, nil }
func (s *recordingMappingStore) DeleteRoleMapping(id int) (db.RoleMapping, []db.MappedRole, error) {
	return db.RoleMapping{}, nil, nil
}
func (s *recordingMappingStore) SyncMappedRoles(id int, provider string, groups []string) ([]db.MappedRole, []db.MappedRole, error) {
	s.calls++
	s.userID, s.provider, s.groups = id, provider, groups
	return nil, nil, s.err
}

func TestKeycloakRoleSync(t *testing.T) {
	store := &recordingMappingStore{}
	svc := NewRoleSyncService(store)
	if _, _, err := svc.SyncLoginRoles(42, "keycloak", map[string]any{"groups": []any{"/ops", "/ops"}}); err != nil {
		t.Fatal(err)
	}
	if store.calls != 1 || store.userID != 42 || store.provider != "keycloak" || !reflect.DeepEqual(store.groups, []string{"/ops"}) {
		t.Fatalf("unexpected role-sync invocation: %+v", store)
	}
	// An explicit empty list must reach persistence to revoke prior mapped roles.
	if _, _, err := svc.SyncLoginRoles(42, "keycloak", map[string]any{"groups": []any{}}); err != nil {
		t.Fatal(err)
	}
	if store.calls != 2 || len(store.groups) != 0 {
		t.Fatalf("explicit empty groups did not reach persistence: %+v", store)
	}
	// Missing claims and password logins cannot silently revoke roles.
	if _, _, err := svc.SyncLoginRoles(42, "keycloak", map[string]any{}); err != nil { t.Fatal(err) }
	if _, _, err := svc.SyncLoginRoles(42, "", nil); err != nil { t.Fatal(err) }
	if store.calls != 2 { t.Fatal("unexpected sync for absent claims or password login") }
	if _, _, err := svc.SyncLoginRoles(42, "keycloak", map[string]any{"groups": 123}); !errors.Is(err, pro_interfaces.ErrGroupsClaim) {
		t.Fatalf("expected malformed claim rejection: %v", err)
	}
	if store.calls != 2 { t.Fatal("malformed claims reached persistence") }
}

func TestKeycloakRoleSyncUnavailable(t *testing.T) {
	svc := NewRoleSyncService(nil)
	if _, _, err := svc.SyncLoginRoles(1, "keycloak", map[string]any{"groups": []any{"/ops"}}); err == nil {
		t.Fatal("expected unavailable mapping store error")
	}
}

func TestKeycloakRoleSyncStoreFailure(t *testing.T) {
	want := errors.New("database unavailable")
	store := &recordingMappingStore{err: want}
	svc := NewRoleSyncService(store)
	_, _, err := svc.SyncLoginRoles(1, "keycloak", map[string]any{"groups": []any{"/ops"}})
	if !errors.Is(err, want) { t.Fatalf("store error not propagated: %v", err) }
}
