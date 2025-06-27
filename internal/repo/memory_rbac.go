package repo

import (
	"deliverymanagement/internal/model"
	"sync"
)

type InMemoryRoleRepo struct {
	mu    sync.RWMutex
	roles map[uint]*model.Role
	next  uint
}

func NewInMemoryRoleRepo() *InMemoryRoleRepo {
	return &InMemoryRoleRepo{roles: make(map[uint]*model.Role), next: 1}
}
func (r *InMemoryRoleRepo) CreateRole(role *model.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	role.ID = r.next
	r.next++
	r.roles[role.ID] = role
	return nil
}
func (r *InMemoryRoleRepo) GetRole(id uint) (*model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, ok := r.roles[id]
	if !ok {
		return nil, nil
	}
	return role, nil
}
func (r *InMemoryRoleRepo) ListRoles() ([]model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []model.Role
	for _, v := range r.roles {
		out = append(out, *v)
	}
	return out, nil
}
func (r *InMemoryRoleRepo) DeleteRole(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, id)
	return nil
}

type InMemoryPermissionRepo struct {
	mu    sync.RWMutex
	perms map[uint]*model.Permission
	next  uint
}

func NewInMemoryPermissionRepo() *InMemoryPermissionRepo {
	return &InMemoryPermissionRepo{perms: make(map[uint]*model.Permission), next: 1}
}
func (r *InMemoryPermissionRepo) CreatePermission(p *model.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.ID = r.next
	r.next++
	r.perms[p.ID] = p
	return nil
}
func (r *InMemoryPermissionRepo) ListPermissions() ([]model.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []model.Permission
	for _, v := range r.perms {
		out = append(out, *v)
	}
	return out, nil
}

type InMemoryRolePermissionRepo struct {
	mu        sync.RWMutex
	rolePerms map[uint]map[uint]struct{}
}

func NewInMemoryRolePermissionRepo() *InMemoryRolePermissionRepo {
	return &InMemoryRolePermissionRepo{rolePerms: make(map[uint]map[uint]struct{})}
}
func (r *InMemoryRolePermissionRepo) AssignPermission(roleID, permID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.rolePerms[roleID] == nil {
		r.rolePerms[roleID] = make(map[uint]struct{})
	}
	r.rolePerms[roleID][permID] = struct{}{}
	return nil
}
func (r *InMemoryRolePermissionRepo) GetPermissions(roleID uint) ([]model.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	permIDs := r.rolePerms[roleID]
	var perms []model.Permission
	for pid := range permIDs {
		perms = append(perms, model.Permission{ID: pid, Name: "manage_roles"}) // For demo, always return manage_roles
	}
	return perms, nil
}

type InMemoryAuditLogRepo struct {
	mu   sync.RWMutex
	logs []model.AuditLog
}

func NewInMemoryAuditLogRepo() *InMemoryAuditLogRepo {
	return &InMemoryAuditLogRepo{}
}
func (r *InMemoryAuditLogRepo) CreateAudit(log *model.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, *log)
	return nil
}
func (r *InMemoryAuditLogRepo) ListAuditLogs() ([]model.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.AuditLog(nil), r.logs...), nil
}
