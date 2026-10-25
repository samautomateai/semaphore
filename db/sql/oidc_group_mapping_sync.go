package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// SyncMappedRoles reconciles a single user's derived assignments for one
// authenticated OIDC provider. Manual roles are never modified.
func (s OIDCGroupMappingStore) SyncMappedRoles(userID int, provider string, groups []string) (added, removed []db.MappedRole, err error) {
	if userID <= 0 || provider == "" || strings.TrimSpace(provider) != provider {
		return nil, nil, db.ErrInvalidOperation
	}
	for _, group := range groups {
		if group == "" || strings.TrimSpace(group) != group {
			return nil, nil, db.ErrInvalidOperation
		}
	}
	tx, err := s.DB.Sql().Begin()
	if err != nil { return nil, nil, err }
	defer tx.Rollback() //nolint:errcheck

	// Lock the user row to serialize concurrent sign-ins on PostgreSQL/MySQL.
	// SQLite uses its write transaction locking semantics.
	if _, err = tx.Exec(s.DB.PrepareQuery("update `user` set id=id where id=?"), userID); err != nil {
		return nil, nil, err
	}
	var exists int64
	exists, err = tx.SelectInt(s.DB.PrepareQuery("select count(*) from `user` where id=?"), userID)
	if err != nil { return nil, nil, err }
	if exists != 1 { return nil, nil, db.ErrNotFound }

	var current []db.MappedRole
	_, err = tx.Select(&current, s.DB.PrepareQuery(`
		select mr.project_id, mr.user_id, u.username, mr.mapping_id,
		       m.provider, m.group_name, m.role_id, r.slug as role_slug,
		       r.name as role_name, r.permissions as role_permissions
		from project__user_mapped_role mr
		join group_role_mapping m on m.id=mr.mapping_id
		join role r on r.id=m.role_id
		join ` + "`user`" + ` u on u.id=mr.user_id
		where mr.user_id=? and m.provider=?`), userID, provider)
	if err != nil { return nil, nil, err }

	var candidates []db.RoleMapping
	_, err = tx.Select(&candidates, s.DB.PrepareQuery(`
		select m.* from group_role_mapping m
		join role r on r.id=m.role_id
		where m.provider=? and (r.project_id=m.project_id or r.project_id is null)`), provider)
	if err != nil { return nil, nil, err }
	groupSet := make(map[string]struct{}, len(groups))
	for _, group := range groups { groupSet[group] = struct{}{} }
	want := make(map[int]db.RoleMapping)
	for _, m := range candidates {
		if _, ok := groupSet[m.Group]; ok { want[m.ID] = m }
	}
	have := make(map[int]db.MappedRole)
	for _, m := range current { have[m.MappingID] = m }
	for _, m := range current {
		if _, ok := want[m.MappingID]; ok { continue }
		if _, err = tx.Exec(s.DB.PrepareQuery(`
			delete from project__user_mapped_role where user_id=? and mapping_id=?`), userID, m.MappingID); err != nil {
			return nil, nil, err
		}
		removed = append(removed, m)
	}
	for _, m := range candidates {
		if _, ok := want[m.ID]; !ok { continue }
		if _, ok := have[m.ID]; ok { continue }
		if _, err = tx.Exec(s.DB.PrepareQuery(`
			insert into project__user_mapped_role(project_id,user_id,mapping_id) values(?,?,?)`),
			m.ProjectID, userID, m.ID); err != nil {
			return nil, nil, err
		}
		var mapped db.MappedRole
		err = tx.SelectOne(&mapped, s.DB.PrepareQuery(`
			select mr.project_id, mr.user_id, u.username, mr.mapping_id,
			       gm.provider, gm.group_name, gm.role_id, r.slug as role_slug,
			       r.name as role_name, r.permissions as role_permissions
			from project__user_mapped_role mr
			join group_role_mapping gm on gm.id=mr.mapping_id
			join role r on r.id=gm.role_id
			join ` + "`user`" + ` u on u.id=mr.user_id
			where mr.user_id=? and mr.mapping_id=?`), userID, m.ID)
		if err != nil { return nil, nil, err }
		added = append(added, mapped)
	}
	if err = tx.Commit(); err != nil {
		if errors.Is(err, sql.ErrTxDone) { return nil, nil, fmt.Errorf("role synchronization transaction closed: %w", err) }
		return nil, nil, err
	}
	return added, removed, nil
}
