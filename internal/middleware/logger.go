package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetupLogging configure le middleware de logging
func SetupLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%v | %s | %3d | %13v | %15s | %-7s %#v\n%s",
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	})
}

// RequestID ajoute un ID unique à chaque requête
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("requestId", requestID)
		c.Next()
	}
}

// CorrelationID gère les IDs de corrélation pour le tracing distribué
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		
		c.Header("X-Correlation-ID", correlationID)
		c.Set("correlationId", correlationID)
		c.Next()
	}
}
