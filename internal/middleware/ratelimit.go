package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig configure le rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
}

// clients stocke les limiteurs par IP
var clients = make(map[string]*rate.Limiter)
var clientsMutex sync.Mutex

// SetupRateLimit configure le middleware de rate limiting
func SetupRateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		clientsMutex.Lock()
		limiter, exists := clients[clientIP]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize)
			clients[clientIP] = limiter
		}
		clientsMutex.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Trop de requêtes, veuillez réessayer plus tard",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupClients nettoie les anciens limiteurs (à appeler périodiquement)
func cleanupClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		clientsMutex.Lock()
		// Nettoyer les clients inactifs (implémentation simple)
		// En production, on pourrait tracker la dernière activité
		clients = make(map[string]*rate.Limiter)
		clientsMutex.Unlock()
	}
}

func init() {
	go cleanupClients()
}
