package repo

import "deliverymanagement/internal/model"

type UserRepository interface {
	CreateUser(user *model.User) error
	FindUserByEmail(email string) (*model.User, error)
}

type DeliveryRepository interface {
	CreateDelivery(delivery *model.Delivery) error
	GetDelivery(id uint) (*model.Delivery, error)
}

type ScanEventRepository interface {
	CreateScanEvent(event *model.ScanEvent) error
	ListScanEvents(deliveryID uint) []*model.ScanEvent
}

type DamageReportRepository interface {
	CreateDamageReport(report *model.DamageReport) error
	ListDamageReports(deliveryID uint) []*model.DamageReport
}
