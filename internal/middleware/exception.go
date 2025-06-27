package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GlobalExceptionHandler handles all panics and validation errors centrally.
func GlobalExceptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				c.Abort()
			}
		}()
		c.Next()
		// Handle binding/validation errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				if e.Type == gin.ErrorTypeBind {
					c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
					return
				}
			}
		}
	}
}
