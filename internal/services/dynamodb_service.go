package services

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/edinfamous/historial-blockchain/internal/models"
)

// DynamoDBService gère les interactions avec DynamoDB
type DynamoDBService struct {
	client                     *dynamodb.Client
	historialTableName         string
	eventoTableName           string
	blockchainEventsTableName string
}

// NewDynamoDBService crée une nouvelle instance de DynamoDBService
func NewDynamoDBService(client *dynamodb.Client, historialTableName, eventoTableName, blockchainEventsTableName string) *DynamoDBService {
	return &DynamoDBService{
		client:                     client,
		historialTableName:         historialTableName,
		eventoTableName:           eventoTableName,
		blockchainEventsTableName: blockchainEventsTableName,
	}
}

// GuardarHistorial sauvegarde l'historial de transparence
func (ddb *DynamoDBService) GuardarHistorial(ctx context.Context, historial *models.HistorialTransparencia) error {
	// Convertir vers les attributs DynamoDB
	item, err := attributevalue.MarshalMap(historial)
	if err != nil {
		return fmt.Errorf("erreur marshalling historial: %w", err)
	}

	// Exécuter PutItem
	_, err = ddb.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(ddb.historialTableName),
		Item:      item,
	})
	
	if err != nil {
		return fmt.Errorf("erreur sauvegarde historial: %w", err)
	}

	log.Printf("✅ Historial sauvegardé: %s - %s", historial.IDProducto, historial.Lote)
	return nil
}

// ObtenerHistorial récupère un historial par ID produit et lote
func (ddb *DynamoDBService) ObtenerHistorial(ctx context.Context, idProducto, lote string) (*models.HistorialTransparencia, error) {
	// La table utilise seulement idProducto comme clé primaire
	// Le paramètre lote peut être utilisé pour filtrer après récupération si nécessaire

	key := map[string]types.AttributeValue{
		"idProducto": &types.AttributeValueMemberS{Value: idProducto},
	}

	result, err := ddb.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(ddb.historialTableName),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération historial: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Non trouvé
	}

	var historial models.HistorialTransparencia
	err = attributevalue.UnmarshalMap(result.Item, &historial)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling historial: %w", err)
	}

	// Si un lote spécifique est demandé, vérifier que ça correspond
	if lote != "" && historial.Lote != lote {
		return nil, nil // Lote ne correspond pas
	}

	return &historial, nil
}

// GuardarEvento sauvegarde un événement vérifié
func (ddb *DynamoDBService) GuardarEvento(ctx context.Context, evento *models.EventoVerificado) error {
	// Convertir vers les attributs DynamoDB
	item, err := attributevalue.MarshalMap(evento)
	if err != nil {
		return fmt.Errorf("erreur marshalling événement: %w", err)
	}

	// Exécuter PutItem avec condition pour éviter les doublons
	_, err = ddb.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(ddb.eventoTableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(idEvento)"),
	})
	
	if err != nil {
		// Si l'élément existe déjà, c'est OK (idempotence)
		if _, ok := err.(*types.ConditionalCheckFailedException); ok {
			log.Printf("⚠️ Événement déjà existant (idempotence): %s", evento.IDEvento)
			return nil
		}
		return fmt.Errorf("erreur sauvegarde événement: %w", err)
	}

	log.Printf("✅ Événement sauvegardé: %s", evento.IDEvento)
	return nil
}

// ObtenerEventos récupère tous les événements pour un produit
func (ddb *DynamoDBService) ObtenerEventos(ctx context.Context, idProducto string) ([]models.EventoVerificado, error) {
	result, err := ddb.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(ddb.eventoTableName),
		KeyConditionExpression: aws.String("idProducto = :idProducto"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idProducto": &types.AttributeValueMemberS{Value: idProducto},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération événements: %w", err)
	}

	var eventos []models.EventoVerificado
	err = attributevalue.UnmarshalListOfMaps(result.Items, &eventos)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling événements: %w", err)
	}

	return eventos, nil
}

// ObtenerEvento récupère un événement spécifique
func (ddb *DynamoDBService) ObtenerEvento(ctx context.Context, idProducto, idEvento string) (*models.EventoVerificado, error) {
	key := map[string]types.AttributeValue{
		"idProducto": &types.AttributeValueMemberS{Value: idProducto},
		"idEvento":   &types.AttributeValueMemberS{Value: idEvento},
	}

	result, err := ddb.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(ddb.eventoTableName),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération événement: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Non trouvé
	}

	var evento models.EventoVerificado
	err = attributevalue.UnmarshalMap(result.Item, &evento)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling événement: %w", err)
	}

	return &evento, nil
}

// GuardarTaskStatus sauvegarde le statut d'une tâche
func (ddb *DynamoDBService) GuardarTaskStatus(ctx context.Context, taskStatus *models.TaskStatus) error {
	// On peut utiliser la même table ou une table dédiée pour les tâches
	// Ici on utilise la table historial avec un préfixe spécial
	taskStatus.TaskID = "TASK#" + taskStatus.TaskID
	
	item, err := attributevalue.MarshalMap(taskStatus)
	if err != nil {
		return fmt.Errorf("erreur marshalling task status: %w", err)
	}

	_, err = ddb.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(ddb.historialTableName),
		Item:      item,
	})
	
	if err != nil {
		return fmt.Errorf("erreur sauvegarde task status: %w", err)
	}

	return nil
}

// ObtenerTaskStatus récupère le statut d'une tâche
func (ddb *DynamoDBService) ObtenerTaskStatus(ctx context.Context, taskID string) (*models.TaskStatus, error) {
	key := map[string]types.AttributeValue{
		"idProducto": &types.AttributeValueMemberS{Value: "TASK#" + taskID},
	}

	result, err := ddb.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(ddb.historialTableName),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération task status: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Non trouvé
	}

	var taskStatus models.TaskStatus
	err = attributevalue.UnmarshalMap(result.Item, &taskStatus)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling task status: %w", err)
	}

	// Nettoyer le préfixe
	taskStatus.TaskID = taskStatus.TaskID[5:] // Enlever "TASK#"

	return &taskStatus, nil
}

// ListarHistorialesInconsistentes liste les historiales avec état inconsistant
func (ddb *DynamoDBService) ListarHistorialesInconsistentes(ctx context.Context) ([]models.HistorialTransparencia, error) {
	// Utiliser un GSI sur estadoActual si disponible
	result, err := ddb.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(ddb.historialTableName),
		FilterExpression: aws.String("estadoActual = :estado"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":estado": &types.AttributeValueMemberS{Value: models.EstadoInconsistente},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur scan historiales inconsistants: %w", err)
	}

	var historiales []models.HistorialTransparencia
	err = attributevalue.UnmarshalListOfMaps(result.Items, &historiales)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling historiales: %w", err)
	}

	return historiales, nil
}

// ObtenerEventosBlockchainPorProducto récupère les événements de la table blockcahin_medysupyly pour un produit
func (ddb *DynamoDBService) ObtenerEventosBlockchainPorProducto(ctx context.Context, idProducto string) ([]models.BlockchainEvent, error) {
	result, err := ddb.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(ddb.blockchainEventsTableName),
		FilterExpression: aws.String("idProducto = :idProducto"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idProducto": &types.AttributeValueMemberS{Value: idProducto},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération événements blockchain: %w", err)
	}

	var eventos []models.BlockchainEvent
	err = attributevalue.UnmarshalListOfMaps(result.Items, &eventos)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling événements blockchain: %w", err)
	}

	return eventos, nil
}

// ObtenerTousEventosBlockchain récupère tous les événements de la table blockcahin_medysupyly
func (ddb *DynamoDBService) ObtenerTousEventosBlockchain(ctx context.Context) ([]models.BlockchainEvent, error) {
	result, err := ddb.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(ddb.blockchainEventsTableName),
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tous les événements blockchain: %w", err)
	}

	var eventos []models.BlockchainEvent
	err = attributevalue.UnmarshalListOfMaps(result.Items, &eventos)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshalling tous les événements blockchain: %w", err)
	}

	return eventos, nil
}
