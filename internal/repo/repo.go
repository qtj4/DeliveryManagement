package repo

import (
	"deliverymanagement/internal/model"
	"time"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	FindUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(email string) error
	ListUsers() ([]*model.User, error)
	SetVerificationToken(email, token string) error
	SetResetToken(email, token string, expiry time.Time) error
	VerifyUser(email string) error
}

type DeliveryRepository interface {
	CreateDelivery(delivery *model.Delivery) error
	GetDelivery(id uint) (*model.Delivery, error)
	ListDeliveries() ([]model.Delivery, error)
}

type ScanEventRepository interface {
	CreateScanEvent(event *model.ScanEvent) error
	ListScanEvents(deliveryID uint) []*model.ScanEvent
}

type DamageReportRepository interface {
	CreateDamageReport(report *model.DamageReport) error
	ListDamageReports(deliveryID uint) []*model.DamageReport
}

type RoleRepository interface {
	CreateRole(role *model.Role) error
	GetRole(id uint) (*model.Role, error)
	ListRoles() ([]model.Role, error)
	DeleteRole(id uint) error
}

type PermissionRepository interface {
	CreatePermission(perm *model.Permission) error
	ListPermissions() ([]model.Permission, error)
}

type RolePermissionRepository interface {
	AssignPermission(roleID, permID uint) error
	GetPermissions(roleID uint) ([]model.Permission, error)
}

type AuditLogRepository interface {
	CreateAudit(log *model.AuditLog) error
	ListAuditLogs() ([]model.AuditLog, error)
}
