package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/require"
)

type mappingAPITestStore struct {
	created []db.RoleMapping
	deleted []int
}
func (s *mappingAPITestStore) GetRoleMappings() ([]db.RoleMapping, error) {
	return append([]db.RoleMapping(nil), s.created...), nil
}
func (s *mappingAPITestStore) CreateRoleMapping(m db.RoleMapping) (db.RoleMapping, error) {
	m.ID = len(s.created) + 1
	s.created = append(s.created, m)
	return m, nil
}
func (s *mappingAPITestStore) DeleteRoleMapping(id int) (db.RoleMapping, []db.MappedRole, error) {
	s.deleted = append(s.deleted, id)
	return db.RoleMapping{ID:id}, nil, nil
}
func (s *mappingAPITestStore) SyncMappedRoles(int, string, []string) ([]db.MappedRole, []db.MappedRole, error) {
	return nil, nil, nil
}

func TestRoleMappingsRejectMalformedRequests(t *testing.T) {
	controller := NewRoleMappingsController(&mappingAPITestStore{}, nil)
	for _, body := range []string{
		`{}`,
		`{"provider":"keycloak","group":" ","project_id":1,"role":"viewer"}`,
		`{"provider":"keycloak","group":"/ops","project_id":0,"role":"viewer"}`,
		`{"provider":"keycloak","group":"/ops","project_id":1}`,
		`not json`,
	} {
		t.Run(body, func(t *testing.T) {
			rec := httptest.NewRecorder()
			controller.AddMapping(rec, httptest.NewRequest(http.MethodPost, "/api/role_mappings", bytes.NewBufferString(body)))
			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestRoleMappingsListAndDelete(t *testing.T) {
	store := &mappingAPITestStore{created:[]db.RoleMapping{{ID:1,Provider:"keycloak",Group:"/ops",ProjectID:2}}}
	controller := NewRoleMappingsController(store, nil)
	rec := httptest.NewRecorder()
	controller.GetMappings(rec, httptest.NewRequest(http.MethodGet, "/api/role_mappings", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var mappings []db.RoleMapping
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &mappings))
	require.Len(t, mappings, 1)
	for _, id := range []string{"bad","0","-1"} {
		rec = httptest.NewRecorder()
		req := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/api/role_mappings/"+id, nil), map[string]string{"mapping_id":id})
		controller.DeleteMapping(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	}
	rec = httptest.NewRecorder()
	req := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/api/role_mappings/1", nil), map[string]string{"mapping_id":"1"})
	controller.DeleteMapping(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, []int{1}, store.deleted)
}
