package server

import (
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// parseOIDCGroups extracts the Keycloak groups claim without coercing arbitrary
// values to strings. A missing claim is distinct from an empty group list:
// callers must not interpret an omitted claim as permission revocation.
func parseOIDCGroups(claims map[string]any) ([]string, bool, error) {
	if claims == nil {
		return nil, false, nil
	}
	value, present := claims["groups"]
	if !present {
		return nil, false, nil
	}
	raw, ok := value.([]any)
	if !ok {
		// Tests and internal callers may provide a typed []string.
		typed, typedOK := value.([]string)
		if !typedOK {
			return nil, true, pro_interfaces.ErrGroupsClaim
		}
		raw = make([]any, len(typed))
		for i, group := range typed {
			raw[i] = group
		}
	}
	groups := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		group, ok := value.(string)
		if !ok || group == "" || strings.TrimSpace(group) != group {
			return nil, true, fmt.Errorf("%w: expected nonempty canonical group string", pro_interfaces.ErrGroupsClaim)
		}
		if _, exists := seen[group]; exists {
			continue
		}
		seen[group] = struct{}{}
		groups = append(groups, group)
	}
	return groups, true, nil
}
