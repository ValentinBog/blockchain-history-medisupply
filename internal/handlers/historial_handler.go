package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/edinfamous/historial-blockchain/internal/models"
	"github.com/edinfamous/historial-blockchain/internal/services"
)

// HistorialHandler gère les requêtes HTTP pour les historiales
type HistorialHandler struct {
	historialService *services.HistorialService
}

// NewHistorialHandler crée une nouvelle instance de HistorialHandler
func NewHistorialHandler(historialService *services.HistorialService) *HistorialHandler {
	return &HistorialHandler{
		historialService: historialService,
	}
}

// ObtenerHistorial maneja GET /api/historial/{idProducto}
func (h *HistorialHandler) ObtenerHistorial(c *gin.Context) {
	idProducto := c.Param("idProducto")
	lote := c.Query("lote")
	full := c.Query("full") == "true"

	if idProducto == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "idProducto est requis",
		})
		return
	}

	historial, err := h.historialService.ObtenerHistorial(c.Request.Context(), idProducto, lote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur récupération historial",
			"details": err.Error(),
		})
		return
	}

	if historial == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Historial non trouvé",
		})
		return
	}

	// Si full=true, inclure les événements détaillés
	if full {
		// Ici on pourrait charger les événements détaillés
		// Pour l'instant, on retourne l'historial tel quel
	}

	c.JSON(http.StatusOK, historial)
}

// ReconstruirHistorial maneja POST /api/historial/reconstruir
func (h *HistorialHandler) ReconstruirHistorial(c *gin.Context) {
	var req models.ReconstruirRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Données invalides",
			"details": err.Error(),
		})
		return
	}

	// Vérifier si on doit traiter en asynchrone ou synchrone
	asyncParam := c.Query("async")
	isAsync := asyncParam == "true" || asyncParam == "1"

	if isAsync {
		// Traitement asynchrone
		taskID, err := h.historialService.ReconstruirHistorialAsync(
			c.Request.Context(), 
			req.IDProducto, 
			req.Lote, 
			req.Force,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Erreur déclenchement reconstruction",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusAccepted, models.ReconstruirResponse{
			Status: "processing",
			TaskID: taskID,
		})
	} else {
		// Traitement synchrone
		historial, err := h.historialService.ReconstruirHistorial(
			c.Request.Context(), 
			req.IDProducto, 
			req.Lote, 
			req.Force,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Erreur reconstruction",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.ReconstruirResponse{
			Status: "completed",
			Data:   historial,
		})
	}
}

// VerificarEvento maneja GET /api/historial/{idProducto}/verify/{idEvento}
func (h *HistorialHandler) VerificarEvento(c *gin.Context) {
	idProducto := c.Param("idProducto")
	idEvento := c.Param("idEvento")

	if idProducto == "" || idEvento == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "idProducto et idEvento sont requis",
		})
		return
	}

	evento, err := h.historialService.VerificarEvento(c.Request.Context(), idProducto, idEvento)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur vérification événement",
			"details": err.Error(),
		})
		return
	}

	if evento == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Événement non trouvé",
		})
		return
	}

	c.JSON(http.StatusOK, evento)
}

// ObtenerEventos maneja GET /api/historial/{idProducto}/events
func (h *HistorialHandler) ObtenerEventos(c *gin.Context) {
	idProducto := c.Param("idProducto")
	tipoEvento := c.Query("tipo")
	
	if idProducto == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "idProducto est requis",
		})
		return
	}

	// Paramètres de pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Obtenir les événements via le service
	eventos, err := h.historialService.ObtenerEventosPorProducto(c.Request.Context(), idProducto, tipoEvento, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur récupération événements",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventos": eventos,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(eventos),
		},
	})
}

// ObtenerStatusTarea maneja GET /api/historial/tasks/{taskId}
func (h *HistorialHandler) ObtenerStatusTarea(c *gin.Context) {
	taskID := c.Param("taskId")

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "taskId est requis",
		})
		return
	}

	taskStatus, err := h.historialService.ObtenerTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur récupération statut tâche",
			"details": err.Error(),
		})
		return
	}

	if taskStatus == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tâche non trouvée",
		})
		return
	}

	c.JSON(http.StatusOK, taskStatus)
}

// ListarInconsistencias maneja GET /api/historial/inconsistencies
func (h *HistorialHandler) ListarInconsistencias(c *gin.Context) {
	// Paramètres de pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	severidad := c.Query("severidad")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Obtenir les inconsistances via le service
	inconsistencias, err := h.historialService.ListarInconsistencias(c.Request.Context(), severidad, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur récupération inconsistances",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inconsistencias": inconsistencias,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(inconsistencias),
		},
		"filtres": gin.H{
			"severidad": severidad,
		},
	})
}
