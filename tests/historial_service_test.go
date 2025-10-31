package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edinfamous/historial-blockchain/internal/models"
	"github.com/edinfamous/historial-blockchain/internal/services"
)

// MockDynamoDBService est un mock pour DynamoDBService
type MockDynamoDBService struct {
	mock.Mock
}

func (m *MockDynamoDBService) GuardarHistorial(ctx context.Context, historial *models.HistorialTransparencia) error {
	args := m.Called(ctx, historial)
	return args.Error(0)
}

func (m *MockDynamoDBService) ObtenerHistorial(ctx context.Context, idProducto, lote string) (*models.HistorialTransparencia, error) {
	args := m.Called(ctx, idProducto, lote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HistorialTransparencia), args.Error(1)
}

func (m *MockDynamoDBService) GuardarEvento(ctx context.Context, evento *models.EventoVerificado) error {
	args := m.Called(ctx, evento)
	return args.Error(0)
}

func (m *MockDynamoDBService) ObtenerEventos(ctx context.Context, idProducto string) ([]models.EventoVerificado, error) {
	args := m.Called(ctx, idProducto)
	return args.Get(0).([]models.EventoVerificado), args.Error(1)
}

func (m *MockDynamoDBService) ObtenerEvento(ctx context.Context, idProducto, idEvento string) (*models.EventoVerificado, error) {
	args := m.Called(ctx, idProducto, idEvento)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EventoVerificado), args.Error(1)
}

func (m *MockDynamoDBService) GuardarTaskStatus(ctx context.Context, taskStatus *models.TaskStatus) error {
	args := m.Called(ctx, taskStatus)
	return args.Error(0)
}

func (m *MockDynamoDBService) ObtenerTaskStatus(ctx context.Context, taskID string) (*models.TaskStatus, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskStatus), args.Error(1)
}

func (m *MockDynamoDBService) ListarHistorialesInconsistentes(ctx context.Context) ([]models.HistorialTransparencia, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.HistorialTransparencia), args.Error(1)
}

// MockBlockchainService est un mock pour BlockchainService
type MockBlockchainService struct {
	mock.Mock
}

func (m *MockBlockchainService) VerificarIntegridad(ctx context.Context, evento *models.EventoVerificado) error {
	args := m.Called(ctx, evento)
	return args.Error(0)
}

// MockKafkaService est un mock pour KafkaService
type MockKafkaService struct {
	mock.Mock
}

func (m *MockKafkaService) PublishHistorialReconstruido(ctx context.Context, event *models.HistorialReconstruidoEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockKafkaService) PublishInconsistencia(ctx context.Context, event *models.InconsistenciaEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Test de base pour HistorialService
func TestHistorialService_TraiterEvenementTransaccion(t *testing.T) {
	// Arrange
	mockDynamoDB := new(MockDynamoDBService)
	mockBlockchain := new(MockBlockchainService)
	mockKafka := new(MockKafkaService)

	service := services.NewHistorialService(
		mockDynamoDB,
		mockBlockchain,
		mockKafka,
		false, // Pas de vérification stricte pour ce test
	)

	event := &models.TransaccionBlockchainEvent{
		SchemaVersion:       "1.0",
		IDEvento:            "test-event-123",
		TipoEvento:          "Ingreso",
		IDProducto:          "prod-test-001",
		Lote:                "lot-2025-01",
		FechaEvento:         time.Now(),
		DatosEvento:         map[string]interface{}{"cantidad": 100},
		HashEvento:          "0xabc123",
		DireccionBlockchain: "0xtxhash",
		ActorEmisor:         "TestProvider",
	}

	// Mock expectations
	mockDynamoDB.On("GuardarEvento", mock.Anything, mock.MatchedBy(func(evento *models.EventoVerificado) bool {
		return evento.IDEvento == "test-event-123" && evento.IDProducto == "prod-test-001"
	})).Return(nil)

	// Act
	err := service.TraiterEvenementTransaccion(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockDynamoDB.AssertExpectations(t)
}

func TestHistorialService_ObtenerHistorial(t *testing.T) {
	// Arrange
	mockDynamoDB := new(MockDynamoDBService)
	mockBlockchain := new(MockBlockchainService)
	mockKafka := new(MockKafkaService)

	service := services.NewHistorialService(
		mockDynamoDB,
		mockBlockchain,
		mockKafka,
		false,
	)

	expectedHistorial := &models.HistorialTransparencia{
		IDProducto:           "prod-test-001",
		Lote:                "lot-2025-01",
		NombreProducto:      "Test Product",
		EstadoActual:        models.EstadoConforme,
		ValidacionBlockchain: true,
		UltimoCheck:         time.Now(),
	}

	// Mock expectations
	mockDynamoDB.On("ObtenerHistorial", mock.Anything, "prod-test-001", "lot-2025-01").Return(expectedHistorial, nil)

	// Act
	result, err := service.ObtenerHistorial(context.Background(), "prod-test-001", "lot-2025-01")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedHistorial, result)
	mockDynamoDB.AssertExpectations(t)
}

func TestHistorialService_ReconstruirHistorial_NoEvents(t *testing.T) {
	// Arrange
	mockDynamoDB := new(MockDynamoDBService)
	mockBlockchain := new(MockBlockchainService)
	mockKafka := new(MockKafkaService)

	service := services.NewHistorialService(
		mockDynamoDB,
		mockBlockchain,
		mockKafka,
		false,
	)

	// Mock expectations
	mockDynamoDB.On("ObtenerHistorial", mock.Anything, "prod-test-001", "lot-2025-01").Return(nil, nil)
	mockDynamoDB.On("ObtenerEventos", mock.Anything, "prod-test-001").Return([]models.EventoVerificado{}, nil)

	// Act
	result, err := service.ReconstruirHistorial(context.Background(), "prod-test-001", "lot-2025-01", false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "aucun événement trouvé")
	mockDynamoDB.AssertExpectations(t)
}
