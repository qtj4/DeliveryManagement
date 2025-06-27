package main

import (
	"deliverymanagement/internal/handler"
	"deliverymanagement/internal/middleware"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"
	"deliverymanagement/pkg/ws"
	"os"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(middleware.GlobalExceptionHandler())
	r.Use(middleware.RateLimiterMiddleware())

	userRepo := repo.NewInMemoryUserRepo()
	deliveryRepo := repo.NewInMemoryDeliveryRepo()
	scanEventRepo := repo.NewInMemoryScanEventRepo()
	damageReportRepo := repo.NewInMemoryDamageReportRepo()
	roleRepo := repo.NewInMemoryRoleRepo()
	permRepo := repo.NewInMemoryPermissionRepo()
	rolePermRepo := repo.NewInMemoryRolePermissionRepo()
	auditRepo := repo.NewInMemoryAuditLogRepo()
	publisher, _ := rabbitmq.New(os.Getenv("RABBITMQ_URL"))
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	hub := ws.NewHub(redisClient)

	authHandler := &handler.AuthHandler{Users: userRepo}
	deliveryHandler := &handler.DeliveryHandler{Deliveries: deliveryRepo, Publisher: publisher, WSHub: hub}
	scanEventHandler := &handler.ScanEventHandler{ScanEvents: scanEventRepo, WSHub: hub}
	damageReportHandler := &handler.DamageReportHandler{DamageReports: damageReportRepo}
	rbacHandler := &handler.RBACHandler{Roles: roleRepo, Perms: permRepo, RolePerms: rolePermRepo, Audit: auditRepo}
	userAdminHandler := &handler.UserAdminHandler{Users: userRepo}
	authFlowHandler := &handler.AuthFlowHandler{Users: userRepo, Publisher: publisher}
	fileHandler := &handler.FileHandler{DamageReports: damageReportRepo}
	notificationRepo := repo.NewInMemoryNotificationRepo()
	analyticsHandler := &handler.AnalyticsHandler{Deliveries: deliveryRepo, Users: userRepo}
	notificationHandler := &handler.NotificationHandler{Notifications: notificationRepo, WSHub: hub}

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		// Password reset & verification
		auth.POST("/reset-request", authFlowHandler.ResetRequest)
		auth.POST("/reset/:token", authFlowHandler.ResetPassword)
		auth.POST("/verify/:token", authFlowHandler.VerifyEmail)
	}

	deliveries := r.Group("/api/deliveries")
	deliveries.Use(handler.JWTAuthMiddleware([]byte("supersecret")))
	{
		deliveries.POST("", deliveryHandler.CreateDelivery)
		deliveries.GET(":id", deliveryHandler.GetDelivery)
		deliveries.POST(":id/assign", handler.DispatcherOnly(), deliveryHandler.AssignDelivery)
		deliveries.GET("/export", deliveryHandler.ExportDeliveries)
	}

	r.POST("/api/scan", scanEventHandler.CreateScanEvent)
	r.POST("/api/damage-report", handler.JWTAuthMiddleware([]byte("supersecret")), handler.CourierOrWarehouseOnly(), damageReportHandler.CreateDamageReport)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	admin := r.Group("/api/admin")
	{
		admin.POST("/roles", rbacHandler.CreateRole)
		admin.GET("/roles", rbacHandler.ListRoles)
		admin.DELETE("/roles/:id", rbacHandler.DeleteRole)
		admin.POST("/permissions", rbacHandler.CreatePermission)
		admin.GET("/permissions", rbacHandler.ListPermissions)
		admin.POST("/role-permissions", rbacHandler.AssignPermission)
		admin.GET("/audit", rbacHandler.ListAuditLogs)

		// User management endpoints
		admin.GET("/users", userAdminHandler.ListUsers)
		admin.POST("/users", userAdminHandler.CreateUser)
		admin.GET("/users/:id", userAdminHandler.GetUser)
		admin.PUT("/users/:id", userAdminHandler.UpdateUser)
		admin.DELETE("/users/:id", userAdminHandler.DeleteUser)
	}

	r.GET("/files/:filename", handler.JWTAuthMiddleware([]byte("supersecret")), fileHandler.ServeFile)

	r.GET("/api/admin/analytics/summary", analyticsHandler.Summary)
	r.GET("/api/admin/analytics/by-courier", analyticsHandler.ByCourier)
	r.GET("/api/notifications", notificationHandler.List)
	r.POST("/api/notifications/:id/read", notificationHandler.MarkRead)

	r.Run() // listen and serve on 0.0.0.0:8080
}
