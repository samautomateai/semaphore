package sql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// OIDCGroupMappingStore uses the tables already installed by migration 2.20.14.
// It is independent of Pro packages and never writes manual project roles.
type OIDCGroupMappingStore struct { DB *SqlDb }

func (s OIDCGroupMappingStore) GetRoleMappings() ([]db.RoleMapping, error) {
	var mappings []db.RoleMapping
	_, err := s.DB.selectAll(&mappings, `
		select m.*, p.name as project_name, r.slug as role_slug, r.name as role_name
		from group_role_mapping m
		join project p on p.id=m.project_id
		join role r on r.id=m.role_id
		order by m.provider, m.group_name, p.name, r.name`)
	return mappings, err
}

func (s OIDCGroupMappingStore) CreateRoleMapping(m db.RoleMapping) (db.RoleMapping, error) {
	if strings.TrimSpace(m.Provider) != m.Provider || m.Provider == "" ||
		strings.TrimSpace(m.Group) != m.Group || m.Group == "" ||
		m.ProjectID <= 0 || m.RoleID <= 0 {
		return db.RoleMapping{}, db.ErrInvalidOperation
	}
	// Role must belong to this project or be a shared role. No global roles.
	var count int64
	err := s.DB.Sql().SelectOne(&count, s.DB.PrepareQuery(`
		select count(*) from role where id=? and (project_id=? or project_id is null)`), m.RoleID, m.ProjectID)
	if err != nil { return db.RoleMapping{}, err }
	if count != 1 { return db.RoleMapping{}, db.ErrInvalidOperation }
	m.Created = tz.Now()
	id, err := s.DB.insert("id", `
		insert into group_role_mapping(provider,group_name,project_id,role_id,created)
		values(?,?,?,?,?)`, m.Provider, m.Group, m.ProjectID, m.RoleID, m.Created)
	if err != nil { return db.RoleMapping{}, err }
	m.ID = id
	return m, nil
}

func (s OIDCGroupMappingStore) DeleteRoleMapping(id int) (db.RoleMapping, []db.MappedRole, error) {
	if id <= 0 { return db.RoleMapping{}, nil, db.ErrInvalidOperation }
	tx, err := s.DB.Sql().Begin()
	if err != nil { return db.RoleMapping{}, nil, err }
	defer tx.Rollback() //nolint:errcheck
	var mapping db.RoleMapping
	err = tx.SelectOne(&mapping, s.DB.PrepareQuery("select * from group_role_mapping where id=?"), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return db.RoleMapping{}, nil, db.ErrNotFound }
		return db.RoleMapping{}, nil, err
	}
	var removed []db.MappedRole
	_, err = tx.Select(&removed, s.DB.PrepareQuery(`
		select mr.project_id, mr.user_id, u.username, mr.mapping_id,
		       m.provider, m.group_name, m.role_id, r.slug as role_slug,
		       r.name as role_name, r.permissions as role_permissions
		from project__user_mapped_role mr
		join group_role_mapping m on m.id=mr.mapping_id
		join role r on r.id=m.role_id
		join ` + "`user`" + ` u on u.id=mr.user_id
		where mr.mapping_id=?`), id)
	if err != nil { return db.RoleMapping{}, nil, err }
	_, err = tx.Exec(s.DB.PrepareQuery("delete from group_role_mapping where id=?"), id)
	if err != nil { return db.RoleMapping{}, nil, err }
	if err = tx.Commit(); err != nil { return db.RoleMapping{}, nil, err }
	return mapping, removed, nil
}
