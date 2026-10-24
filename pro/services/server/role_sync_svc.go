package server

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// RoleSyncServiceImpl delegates to the configured role-mapping store.
// Community's current store is a no-op; persistence must be implemented
// before this provides effective group-derived authorization.
type RoleSyncServiceImpl struct {
	mappings db.RoleMappingStore
}

func NewRoleSyncService(mappings db.RoleMappingStore) pro_interfaces.RoleSyncService {
	return &RoleSyncServiceImpl{mappings: mappings}
}

func (s *RoleSyncServiceImpl) SyncLoginRoles(userID int, provider string, claims map[string]any) ([]db.MappedRole, []db.MappedRole, error) {
	// Password and LDAP sign-ins do not assert OIDC group membership.
	if provider == "" {
		return nil, nil, nil
	}
	groups, present, err := parseOIDCGroups(claims)
	if err != nil {
		return nil, nil, err
	}
	// An omitted claim is not an assertion of empty membership.
	if !present {
		return nil, nil, nil
	}
	if userID <= 0 || s.mappings == nil {
		return nil, nil, errors.New("OIDC group role synchronization unavailable")
	}
	return s.mappings.SyncMappedRoles(userID, provider, groups)
}
