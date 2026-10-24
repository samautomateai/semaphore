package server

import (
	"errors"
	"reflect"
	"testing"

	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func TestParseOIDCGroups(t *testing.T) {
	tests := []struct {
		name string
		claims map[string]any
		want []string
		present bool
		invalid bool
	}{
		{name: "missing", claims: map[string]any{}, present: false},
		{name: "nil claims", claims: nil, present: false},
		{name: "explicitly empty", claims: map[string]any{"groups": []any{}}, want: []string{}, present: true},
		{name: "Keycloak groups", claims: map[string]any{"groups": []any{"/ops", "/deploy"}}, want: []string{"/ops", "/deploy"}, present: true},
		{name: "deduplicate", claims: map[string]any{"groups": []any{"/ops", "/ops"}}, want: []string{"/ops"}, present: true},
		{name: "typed slice", claims: map[string]any{"groups": []string{"/ops"}}, want: []string{"/ops"}, present: true},
		{name: "null", claims: map[string]any{"groups": nil}, present: true, invalid: true},
		{name: "scalar", claims: map[string]any{"groups": "/ops"}, present: true, invalid: true},
		{name: "mixed", claims: map[string]any{"groups": []any{"/ops", 4}}, present: true, invalid: true},
		{name: "blank", claims: map[string]any{"groups": []any{""}}, present: true, invalid: true},
		{name: "whitespace", claims: map[string]any{"groups": []any{" /ops "}}, present: true, invalid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, present, err := parseOIDCGroups(tc.claims)
			if present != tc.present {
				t.Fatalf("present=%v, want %v", present, tc.present)
			}
			if tc.invalid {
				if !errors.Is(err, pro_interfaces.ErrGroupsClaim) {
					t.Fatalf("error=%v, want ErrGroupsClaim", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("groups=%v, want %v", got, tc.want)
			}
		})
	}
}
