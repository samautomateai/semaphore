package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// RoleSyncServiceImpl is the community stub: group role mappings are a Pro feature.
type RoleSyncServiceImpl struct{}

func NewRoleSyncService(mappings db.RoleMappingStore) pro_interfaces.RoleSyncService {
	return &RoleSyncServiceImpl{}
}

func (s *RoleSyncServiceImpl) SyncLoginRoles(userID int, provider string, claims map[string]any) ([]db.MappedRole, []db.MappedRole, error) {
	// Reject malformed OIDC group claims even while Community group-role
	// persistence is unavailable. Never imply that roles were synchronized.
	if provider != "" {
		if _, _, err := parseOIDCGroups(claims); err != nil {
			return nil, nil, err
		}
	}
	return nil, nil, nil
}
