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
	authHandler := &handler.AuthHandler{Users: userRepo}
	deliveryHandler := &handler.DeliveryHandler{Deliveries: deliveryRepo}

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
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
