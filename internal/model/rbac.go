package model

type Role struct {
	ID   uint
	Name string
}

type Permission struct {
	ID   uint
	Name string
}

type RolePermission struct {
	RoleID       uint
	PermissionID uint
}

type AuditLog struct {
	ID        uint
	UserID    uint
	Action    string
	Resource  string
	Success   bool
	Timestamp int64
}
