package main

import (
	"deliverymanagement/internal/handler"
	"deliverymanagement/internal/repo"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	userRepo := repo.NewInMemoryUserRepo()
	deliveryRepo := repo.NewInMemoryDeliveryRepo()
	scanEventRepo := repo.NewInMemoryScanEventRepo()
	damageReportRepo := repo.NewInMemoryDamageReportRepo()
	authHandler := &handler.AuthHandler{Users: userRepo}
	deliveryHandler := &handler.DeliveryHandler{Deliveries: deliveryRepo}
	scanEventHandler := &handler.ScanEventHandler{ScanEvents: scanEventRepo}
	damageReportHandler := &handler.DamageReportHandler{DamageReports: damageReportRepo}

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	deliveries := r.Group("/api/deliveries")
	deliveries.Use(handler.JWTAuthMiddleware([]byte("supersecret")))
	{
		deliveries.POST("", deliveryHandler.CreateDelivery)
		deliveries.GET(":id", deliveryHandler.GetDelivery)
		deliveries.POST(":id/assign", handler.DispatcherOnly(), deliveryHandler.AssignDelivery)
	}

	r.POST("/api/scan", handler.JWTAuthMiddleware([]byte("supersecret")), handler.CourierOrWarehouseOnly(), scanEventHandler.CreateScanEvent)
	r.POST("/api/damage-report", handler.JWTAuthMiddleware([]byte("supersecret")), handler.CourierOrWarehouseOnly(), damageReportHandler.CreateDamageReport)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
