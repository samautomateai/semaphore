package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// Route-level system permissions are enforced by the parent router.
type RoleMappingsController struct {
	mappings db.RoleMappingStore
	store db.Store
}

func NewRoleMappingsController(mappings db.RoleMappingStore, store db.Store) *RoleMappingsController {
	return &RoleMappingsController{mappings: mappings, store: store}
}

func (c *RoleMappingsController) GetMappings(w http.ResponseWriter, r *http.Request) {
	mappings, err := c.mappings.GetRoleMappings()
	if err != nil { helpers.WriteError(w, err); return }
	if mappings == nil { mappings = []db.RoleMapping{} }
	helpers.WriteJSON(w, http.StatusOK, mappings)
}

func (c *RoleMappingsController) AddMapping(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Provider string `json:"provider"`
		Group string `json:"group"`
		ProjectID int `json:"project_id"`
		Role string `json:"role"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid mapping request", http.StatusBadRequest); return
	}
	if request.Provider == "" || strings.TrimSpace(request.Provider) != request.Provider ||
		request.Group == "" || strings.TrimSpace(request.Group) != request.Group ||
		request.ProjectID <= 0 || request.Role == "" {
		http.Error(w, "invalid mapping fields", http.StatusBadRequest); return
	}
	role, err := c.store.GetProjectOrSharedRoleBySlug(request.ProjectID, request.Role)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) { http.Error(w, "role not found", http.StatusBadRequest); return }
		helpers.WriteError(w, err); return
	}
	mapping, err := c.mappings.CreateRoleMapping(db.RoleMapping{
		Provider: request.Provider, Group: request.Group,
		ProjectID: request.ProjectID, RoleID: role.ID,
	})
	if err != nil { helpers.WriteError(w, err); return }
	mapping.RoleSlug = role.Slug
	mapping.RoleName = role.Name
	helpers.WriteJSON(w, http.StatusCreated, mapping)
}

func (c *RoleMappingsController) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["mapping_id"])
	if err != nil || id <= 0 { http.Error(w, "invalid mapping id", http.StatusBadRequest); return }
	_, _, err = c.mappings.DeleteRoleMapping(id)
	if err != nil { helpers.WriteError(w, err); return }
	w.WriteHeader(http.StatusNoContent)
}
