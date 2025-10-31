package models

import "time"

// HistorialTransparencia représente l'agrégat racine
type HistorialTransparencia struct {
	IDProducto           string             `json:"idProducto" dynamodbav:"idProducto"`
	Lote                string             `json:"lote" dynamodbav:"lote"`
	NombreProducto      string             `json:"nombreProducto" dynamodbav:"nombreProducto"`
	Fabricante          string             `json:"fabricante" dynamodbav:"fabricante"`
	EstadoActual        string             `json:"estadoActual" dynamodbav:"estadoActual"` // Conforme/Inconsistente/Partiel
	ValidacionBlockchain bool              `json:"validacionBlockchain" dynamodbav:"validacionBlockchain"`
	UltimoCheck        time.Time          `json:"ultimoCheck" dynamodbav:"ultimoCheck"`
	RawPayload          string             `json:"rawPayload,omitempty" dynamodbav:"rawPayload"`
	Metadata            map[string]string  `json:"metadata" dynamodbav:"metadata"`
	CreatedAt           time.Time          `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt" dynamodbav:"updatedAt"`
}

// EventoVerificado représente un événement vérifié
type EventoVerificado struct {
	IDProducto            string            `json:"idProducto" dynamodbav:"idProducto"`
	IDEvento              string            `json:"idEvento" dynamodbav:"idEvento"`
	TipoEvento            string            `json:"tipoEvento" dynamodbav:"tipoEvento"`
	Fecha                 time.Time         `json:"fecha" dynamodbav:"fecha"`
	Ubicacion             string            `json:"ubicacion" dynamodbav:"ubicacion"`
	DatosEvento           map[string]interface{} `json:"datosEvento" dynamodbav:"datosEvento"`
	HashEvento            string            `json:"hashEvento" dynamodbav:"hashEvento"`
	ReferenciaBlockchain  string            `json:"referenciaBlockchain" dynamodbav:"referenciaBlockchain"`
	ResultadoVerificacion string            `json:"resultadoVerificacion" dynamodbav:"resultadoVerificacion"`
	Observaciones         string            `json:"observaciones" dynamodbav:"observaciones"`
	RawPayload            string            `json:"rawPayload" dynamodbav:"rawPayload"`
	CreatedAt            time.Time         `json:"createdAt" dynamodbav:"createdAt"`
}

// TransaccionBlockchainEvent représente l'événement reçu de TransaccionBlockchain
type TransaccionBlockchainEvent struct {
	SchemaVersion        string            `json:"schemaVersion"`
	IDEvento             string            `json:"idEvento"`
	TipoEvento           string            `json:"tipoEvento"`
	IDProducto           string            `json:"idProducto"`
	Lote                 string            `json:"lote"`
	FechaEvento          time.Time         `json:"fechaEvento"`
	DatosEvento          map[string]interface{} `json:"datosEvento"`
	HashEvento           string            `json:"hashEvento"`
	DireccionBlockchain  string            `json:"direccionBlockchain"`
	ActorEmisor          string            `json:"actorEmisor"`
	FirmaDigital         string            `json:"firmaDigital"`
	Metadatos            map[string]interface{} `json:"metadatos"`
}

// HistorialReconstruidoEvent représente l'événement émis après reconstruction
type HistorialReconstruidoEvent struct {
	SchemaVersion      string             `json:"schemaVersion"`
	IDProducto         string             `json:"idProducto"`
	Lote              string             `json:"lote"`
	Estado            string             `json:"estado"`
	EventosVerificados []EventoVerificado `json:"eventosVerificados"`
	Timestamp         time.Time          `json:"timestamp"`
	CorrelationID     string             `json:"correlationId"`
}

// InconsistenciaEvent représente l'événement émis lors d'une inconsistance
type InconsistenciaEvent struct {
	SchemaVersion string                 `json:"schemaVersion"`
	IDProducto    string                 `json:"idProducto"`
	Lote         string                 `json:"lote"`
	Detalles     []InconsistenciaDetalle `json:"detalles"`
	Timestamp    time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlationId"`
}

// InconsistenciaDetalle détaille une inconsistance
type InconsistenciaDetalle struct {
	IDEvento string `json:"idEvento"`
	Error    string `json:"error"`
}

// ReconstruirRequest représente la requête de reconstruction
type ReconstruirRequest struct {
	IDProducto string `json:"idProducto" validate:"required"`
	Lote      string `json:"lote"`
	Force     bool   `json:"force"`
}

// ReconstruirResponse représente la réponse de reconstruction
type ReconstruirResponse struct {
	Status string `json:"status"`
	TaskID string `json:"taskId,omitempty"`
	Data   *HistorialTransparencia `json:"data,omitempty"`
}

// HashCriptografico value object
type HashCriptografico struct {
	Algoritmo  string `json:"algoritmo"`
	ValorHash  string `json:"valorHash"`
}

// VerificarIntegridad vérifie l'intégrité du hash
func (h *HashCriptografico) VerificarIntegridad(valorLocal string) bool {
	return h.ValorHash == valorLocal
}

// FirmaDigital value object
type FirmaDigital struct {
	Certificado string `json:"certificado"`
	Firma       string `json:"firma"`
}

// CondicionTransporte value object
type CondicionTransporte struct {
	Temperatura     string            `json:"temperatura"`
	Humedad        string            `json:"humedad"`
	RangoPermitido map[string]string `json:"rangoPermitido"`
}

// TaskStatus représente le statut d'une tâche de reconstruction
type TaskStatus struct {
	TaskID    string    `json:"taskId"`
	Status    string    `json:"status"` // processing, completed, failed
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Constantes pour les résultats de vérification
const (
	VerificacionOK           = "OK"
	VerificacionHashMismatch = "HASH_MISMATCH" 
	VerificacionFirmaInvalida = "FIRMA_INVALIDA"
	VerificacionNotFound     = "NOT_FOUND"
)

// Constantes pour les états
const (
	EstadoConforme      = "Conforme"
	EstadoInconsistente = "Inconsistente"
	EstadoPartiel       = "Partiel"
)

// Constantes pour les statuts de tâches
const (
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// Inconsistencia représente une inconsistance détectée
type Inconsistencia struct {
	ID             string `json:"id" dynamodbav:"id"`
	IDProducto     string `json:"idProducto" dynamodbav:"idProducto"`
	IDEvento       string `json:"idEvento" dynamodbav:"idEvento"`
	Tipo           string `json:"tipo" dynamodbav:"tipo"`
	Severidad      string `json:"severidad" dynamodbav:"severidad"`
	Descripcion    string `json:"descripcion" dynamodbav:"descripcion"`
	FechaDeteccion string `json:"fechaDeteccion" dynamodbav:"fechaDeteccion"`
	Resolu         bool   `json:"resolu" dynamodbav:"resolu"`
}
