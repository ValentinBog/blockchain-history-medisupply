package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler gère les endpoints de santé
type HealthHandler struct{}

// NewHealthHandler crée une nouvelle instance de HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck maneja GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "historial-blockchain",
		"version": "1.0.0",
	})
}

// ReadinessCheck maneja GET /health/ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Ici on pourrait vérifier les dépendances (DB, Kafka, Blockchain)
	// Pour l'instant, on retourne toujours OK
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"dependencies": gin.H{
			"database":   "ok",
			"kafka":      "ok",
			"blockchain": "ok",
		},
	})
}

// LivenessCheck maneja GET /health/live
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
