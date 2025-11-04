package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/edinfamous/historial-blockchain/internal/models"
)

// HistorialService orchestre la reconstruction et vérification des historiales
type HistorialService struct {
	dynamoDBService   *DynamoDBService
	blockchainService *BlockchainService
	kafkaService      *KafkaService
	strictVerification bool
}

// NewHistorialService crée une nouvelle instance de HistorialService
func NewHistorialService(
	dynamoDBService *DynamoDBService,
	blockchainService *BlockchainService,
	kafkaService *KafkaService,
	strictVerification bool,
) *HistorialService {
	return &HistorialService{
		dynamoDBService:   dynamoDBService,
		blockchainService: blockchainService,
		kafkaService:      kafkaService,
		strictVerification: strictVerification,
	}
}

// ReconstruirHistorial reconstruit l'historial complet d'un produit
func (hs *HistorialService) ReconstruirHistorial(ctx context.Context, idProducto, lote string, force bool) (*models.HistorialTransparencia, error) {
	log.Printf("🔄 Début reconstruction historial: %s - %s", idProducto, lote)

	// Vérifier si l'historial existe déjà et n'est pas forcé
	if !force {
		existingHistorial, err := hs.dynamoDBService.ObtenerHistorial(ctx, idProducto, lote)
		if err != nil {
			return nil, fmt.Errorf("erreur vérification historial existant: %w", err)
		}
		if existingHistorial != nil && time.Since(existingHistorial.UltimoCheck) < time.Hour {
			log.Printf("📋 Historial récent trouvé, retour sans reconstruction")
			return existingHistorial, nil
		}
	}

	// Récupérer tous les événements pour ce produit
	eventos, err := hs.dynamoDBService.ObtenerEventos(ctx, idProducto)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération événements: %w", err)
	}

	if len(eventos) == 0 {
		return nil, fmt.Errorf("aucun événement trouvé pour le produit %s", idProducto)
	}

	// Vérifier chaque événement
	eventosVerificados := make([]models.EventoVerificado, 0, len(eventos))
	var inconsistencias []models.InconsistenciaDetalle
	
	for _, evento := range eventos {
		// Filtrer par lote si spécifié
		if lote != "" {
			// Le lote pourrait être dans DatosEvento
			if eventoLote, ok := evento.DatosEvento["lote"].(string); ok {
				if eventoLote != lote {
					continue
				}
			}
		}

		// Vérifier l'événement contre la blockchain si strict verification
		if hs.strictVerification && evento.ReferenciaBlockchain != "" {
			err := hs.blockchainService.VerificarIntegridad(ctx, &evento)
			if err != nil {
				log.Printf("⚠️ Échec vérification événement %s: %v", evento.IDEvento, err)
				inconsistencias = append(inconsistencias, models.InconsistenciaDetalle{
					IDEvento: evento.IDEvento,
					Error:    evento.ResultadoVerificacion,
				})
			}
		} else {
			// Si pas de vérification stricte, marquer comme OK
			evento.ResultadoVerificacion = models.VerificacionOK
		}

		eventosVerificados = append(eventosVerificados, evento)
	}

	// Déterminer l'état global
	estadoActual := hs.determinerEstadoGlobal(eventosVerificados)

	// Construire l'historial
	historial := &models.HistorialTransparencia{
		IDProducto:           idProducto,
		Lote:                lote,
		EstadoActual:        estadoActual,
		ValidacionBlockchain: hs.strictVerification && len(inconsistencias) == 0,
		UltimoCheck:         time.Now(),
		Metadata:            make(map[string]string),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Extraire informations des événements
	if len(eventosVerificados) > 0 {
		// Prendre le nom et fabricant du premier événement (ou du plus récent)
		primerEvento := eventosVerificados[0]
		if nombreProducto, ok := primerEvento.DatosEvento["nombreProducto"].(string); ok {
			historial.NombreProducto = nombreProducto
		}
		if fabricante, ok := primerEvento.DatosEvento["fabricante"].(string); ok {
			historial.Fabricante = fabricante
		}
	}

	// Sauvegarder l'historial
	err = hs.dynamoDBService.GuardarHistorial(ctx, historial)
	if err != nil {
		return nil, fmt.Errorf("erreur sauvegarde historial: %w", err)
	}

	// Publier événements selon le résultat
	correlationID := uuid.New().String()
	
	if len(inconsistencias) == 0 {
		// Publier événement de reconstruction réussie
		event := &models.HistorialReconstruidoEvent{
			SchemaVersion:      "1.0",
			IDProducto:         idProducto,
			Lote:              lote,
			Estado:            estadoActual,
			EventosVerificados: eventosVerificados,
			Timestamp:         time.Now(),
			CorrelationID:     correlationID,
		}
		
		if err := hs.kafkaService.PublishHistorialReconstruido(ctx, event); err != nil {
			log.Printf("⚠️ Erreur publication événement reconstruction: %v", err)
		}
	} else {
		// Publier événement d'inconsistance
		event := &models.InconsistenciaEvent{
			SchemaVersion: "1.0",
			IDProducto:    idProducto,
			Lote:         lote,
			Detalles:     inconsistencias,
			Timestamp:    time.Now(),
			CorrelationID: correlationID,
		}
		
		if err := hs.kafkaService.PublishInconsistencia(ctx, event); err != nil {
			log.Printf("⚠️ Erreur publication événement inconsistance: %v", err)
		}
	}

	log.Printf("✅ Reconstruction terminée: %s - %s (État: %s)", idProducto, lote, estadoActual)
	return historial, nil
}

// ReconstruirHistorialAsync lance la reconstruction en asynchrone
func (hs *HistorialService) ReconstruirHistorialAsync(ctx context.Context, idProducto, lote string, force bool) (string, error) {
	taskID := uuid.New().String()

	// Créer le statut de tâche
	taskStatus := &models.TaskStatus{
		TaskID:    taskID,
		Status:    models.TaskStatusProcessing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := hs.dynamoDBService.GuardarTaskStatus(ctx, taskStatus)
	if err != nil {
		return "", fmt.Errorf("erreur création tâche: %w", err)
	}

	// Lancer la reconstruction en arrière-plan
	go func() {
		bgCtx := context.Background()
		
		historial, err := hs.ReconstruirHistorial(bgCtx, idProducto, lote, force)
		
		// Mettre à jour le statut
		taskStatus.UpdatedAt = time.Now()
		if err != nil {
			taskStatus.Status = models.TaskStatusFailed
			taskStatus.Error = err.Error()
		} else {
			taskStatus.Status = models.TaskStatusCompleted
			resultBytes, _ := json.Marshal(historial)
			taskStatus.Result = string(resultBytes)
		}
		
		if err := hs.dynamoDBService.GuardarTaskStatus(bgCtx, taskStatus); err != nil {
			log.Printf("❌ Erreur mise à jour statut tâche: %v", err)
		}
	}()

	return taskID, nil
}

// ObtenerHistorial récupère un historial existant
func (hs *HistorialService) ObtenerHistorial(ctx context.Context, idProducto, lote string) (*models.HistorialTransparencia, error) {
	// Utiliser la vraie base de données DynamoDB
	return hs.dynamoDBService.ObtenerHistorial(ctx, idProducto, lote)
}

// VerificarEvento vérifie un événement spécifique
func (hs *HistorialService) VerificarEvento(ctx context.Context, idProducto, idEvento string) (*models.EventoVerificado, error) {
	// Utiliser la vraie base de données DynamoDB
	evento, err := hs.dynamoDBService.ObtenerEvento(ctx, idProducto, idEvento)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération événement: %w", err)
	}
	
	if evento == nil {
		return nil, fmt.Errorf("événement non trouvé")
	}

	// Vérifier contre la blockchain
	if evento.ReferenciaBlockchain != "" {
		err := hs.blockchainService.VerificarIntegridad(ctx, evento)
		if err != nil {
			log.Printf("⚠️ Échec vérification événement %s: %v", idEvento, err)
		}
		
		// Sauvegarder le résultat de vérification
		err = hs.dynamoDBService.GuardarEvento(ctx, evento)
		if err != nil {
			log.Printf("⚠️ Erreur sauvegarde événement vérifié: %v", err)
		}
	}

	return evento, nil
}

// TraiterEvenementTransaccion traite un événement reçu de TransaccionBlockchain
func (hs *HistorialService) TraiterEvenementTransaccion(ctx context.Context, event *models.TransaccionBlockchainEvent) error {
	log.Printf("🔄 Traitement événement: %s", event.IDEvento)

	// Convertir l'événement en EventoVerificado
	eventoVerificado := &models.EventoVerificado{
		IDProducto:           event.IDProducto,
		IDEvento:             event.IDEvento,
		TipoEvento:           event.TipoEvento,
		Fecha:                event.FechaEvento,
		Ubicacion:            event.ActorEmisor, // Utiliser l'acteur comme ubicacion
		DatosEvento:          event.DatosEvento,
		HashEvento:           event.HashEvento,
		ReferenciaBlockchain: event.DireccionBlockchain,
		ResultadoVerificacion: models.VerificacionOK, // Par défaut, sera vérifié plus tard
		RawPayload:           "", // Pourrait être rempli avec l'événement brut
		CreatedAt:            time.Now(),
	}

	// Ajouter le lote aux données si présent
	if event.Lote != "" {
		eventoVerificado.DatosEvento["lote"] = event.Lote
	}

	// Sauvegarder l'événement (idempotent)
	err := hs.dynamoDBService.GuardarEvento(ctx, eventoVerificado)
	if err != nil {
		return fmt.Errorf("erreur sauvegarde événement: %w", err)
	}

	// Si la vérification stricte est activée, vérifier immédiatement
	if hs.strictVerification && eventoVerificado.ReferenciaBlockchain != "" {
		err := hs.blockchainService.VerificarIntegridad(ctx, eventoVerificado)
		if err != nil {
			log.Printf("⚠️ Échec vérification immédiate événement %s: %v", event.IDEvento, err)
		}
		
		// Re-sauvegarder avec le résultat de vérification
		err = hs.dynamoDBService.GuardarEvento(ctx, eventoVerificado)
		if err != nil {
			log.Printf("⚠️ Erreur sauvegarde événement vérifié: %v", err)
		}
	}

	log.Printf("✅ Événement traité: %s", event.IDEvento)
	return nil
}

// ObtenerTaskStatus récupère le statut d'une tâche
func (hs *HistorialService) ObtenerTaskStatus(ctx context.Context, taskID string) (*models.TaskStatus, error) {
	return hs.dynamoDBService.ObtenerTaskStatus(ctx, taskID)
}

// determinerEstadoGlobal détermine l'état global basé sur les événements vérifiés
func (hs *HistorialService) determinerEstadoGlobal(eventos []models.EventoVerificado) string {
	if len(eventos) == 0 {
		return models.EstadoPartiel
	}

	conforme := 0
	total := len(eventos)

	for _, evento := range eventos {
		if evento.ResultadoVerificacion == models.VerificacionOK {
			conforme++
		}
	}

	if conforme == total {
		return models.EstadoConforme
	} else if conforme == 0 {
		return models.EstadoInconsistente
	} else {
		return models.EstadoPartiel
	}
}

// ObtenerEventosPorProducto récupère les événements d'un produit avec pagination
func (hs *HistorialService) ObtenerEventosPorProducto(ctx context.Context, idProducto, tipoEvento string, page, limit int) ([]models.EventoVerificado, error) {
	// Calculer l'offset pour la pagination
	offset := (page - 1) * limit

	// Pour l'instant, simulons des données d'événements
	// Dans une vraie implémentation, on ferait appel au DynamoDB service
	// Parse dates
	fecha1, _ := time.Parse("2006-01-02T15:04:05Z", "2024-01-15T10:30:00Z")
	fecha2, _ := time.Parse("2006-01-02T15:04:05Z", "2024-02-15T14:20:00Z")

	eventos := []models.EventoVerificado{
		{
			IDEvento:             "EVT456",
			IDProducto:          idProducto,
			TipoEvento:          "INGRESO",
			Fecha:               fecha1,
			ReferenciaBlockchain: "0x123abc...",
			ResultadoVerificacion: "VERIFICADO",
			DatosEvento: map[string]interface{}{
				"cantidad":  100,
				"lote":      "L001",
				"proveedor": "PROV001",
			},
		},
		{
			IDEvento:             "EVT789",
			IDProducto:          idProducto,
			TipoEvento:          "EGRESO",
			Fecha:               fecha2,
			ReferenciaBlockchain: "0x456def...",
			ResultadoVerificacion: "VERIFICADO",
			DatosEvento: map[string]interface{}{
				"cantidad": 50,
				"destino":  "HOSPITAL_001",
			},
		},
	}

	// Filtrer par type d'événement si spécifié
	var eventosFiltrados []models.EventoVerificado
	for _, evento := range eventos {
		if tipoEvento == "" || evento.TipoEvento == tipoEvento {
			eventosFiltrados = append(eventosFiltrados, evento)
		}
	}

	// Appliquer la pagination
	start := offset
	if start >= len(eventosFiltrados) {
		return []models.EventoVerificado{}, nil
	}

	end := start + limit
	if end > len(eventosFiltrados) {
		end = len(eventosFiltrados)
	}

	return eventosFiltrados[start:end], nil
}

// ListarInconsistencias récupère les inconsistances avec filtrage et pagination
func (hs *HistorialService) ListarInconsistencias(ctx context.Context, severidad string, page, limit int) ([]models.Inconsistencia, error) {
	// Utiliser la vraie base de données DynamoDB pour récupérer les inconsistances
	historiales, err := hs.dynamoDBService.ListarHistorialesInconsistentes(ctx)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération inconsistances: %w", err)
	}

	// Convertir les historiales en inconsistances (logique métier à adapter selon vos besoins)
	var inconsistencias []models.Inconsistencia
	for _, historial := range historiales {
		if !historial.ValidacionBlockchain {
			inconsistencias = append(inconsistencias, models.Inconsistencia{
				ID:           fmt.Sprintf("INC_%s", historial.IDProducto),
				IDProducto:   historial.IDProducto,
				IDEvento:     "", // À récupérer depuis les événements si nécessaire
				Tipo:         "VALIDATION_FAILED",
				Severidad:    "ALTA",
				Descripcion:  "Validation blockchain échouée",
				FechaDeteccion: historial.UpdatedAt.Format(time.RFC3339),
				Resolu:       false,
			})
		}
	}

	// Filtrer par sévérité si spécifié
	var inconsistenciasFiltradas []models.Inconsistencia
	for _, inc := range inconsistencias {
		if severidad == "" || inc.Severidad == severidad {
			inconsistenciasFiltradas = append(inconsistenciasFiltradas, inc)
		}
	}

	// Appliquer la pagination
	offset := (page - 1) * limit
	start := offset
	if start >= len(inconsistenciasFiltradas) {
		return []models.Inconsistencia{}, nil
	}

	end := start + limit
	if end > len(inconsistenciasFiltradas) {
		end = len(inconsistenciasFiltradas)
	}

	return inconsistenciasFiltradas[start:end], nil
}
