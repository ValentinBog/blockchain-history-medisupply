package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/edinfamous/historial-blockchain/internal/models"
)

// KafkaService gère les interactions avec Kafka
type KafkaService struct {
	reader          *kafka.Reader
	writer          *kafka.Writer
	bootstrapServers string
	consumerGroup    string
	topic           string
	producerTopic   string
}

// NewKafkaService crée une nouvelle instance de KafkaService
func NewKafkaService(bootstrapServers, consumerGroup, topic, producerTopic string) *KafkaService {
	// Configuration du consumer
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{bootstrapServers},
		Topic:       topic,
		GroupID:     consumerGroup,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
	})

	// Configuration du producer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(bootstrapServers),
		Topic:        producerTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 100 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &KafkaService{
		reader:          reader,
		writer:          writer,
		bootstrapServers: bootstrapServers,
		consumerGroup:    consumerGroup,
		topic:           topic,
		producerTopic:   producerTopic,
	}
}

// ConsumeEvents consomme les événements de TransaccionBlockchain
func (ks *KafkaService) ConsumeEvents(ctx context.Context, handler func(event *models.TransaccionBlockchainEvent) error) error {
	log.Printf("🎧 Début de consommation des événements depuis le topic: %s", ks.topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Arrêt de la consommation d'événements")
			return ctx.Err()
		default:
			// Lire le message suivant
			msg, err := ks.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("❌ Erreur lecture message Kafka: %v", err)
				continue
			}

			log.Printf("📨 Message reçu: partition=%d offset=%d key=%s", 
				msg.Partition, msg.Offset, string(msg.Key))

			// Parser l'événement
			var event models.TransaccionBlockchainEvent
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				log.Printf("❌ Erreur parsing événement: %v", err)
				continue
			}

			// Traiter l'événement
			if err := handler(&event); err != nil {
				log.Printf("❌ Erreur traitement événement %s: %v", event.IDEvento, err)
				// En production, envoyer vers DLQ
				continue
			}

			log.Printf("✅ Événement traité avec succès: %s", event.IDEvento)
		}
	}
}

// PublishHistorialReconstruido publie un événement de reconstruction d'historial
func (ks *KafkaService) PublishHistorialReconstruido(ctx context.Context, event *models.HistorialReconstruidoEvent) error {
	return ks.publishEvent(ctx, "event.historial.reconstruido", event)
}

// PublishInconsistencia publie un événement d'inconsistance
func (ks *KafkaService) PublishInconsistencia(ctx context.Context, event *models.InconsistenciaEvent) error {
	return ks.publishEvent(ctx, "event.historial.inconsistencia", event)
}

// publishEvent publie un événement générique
func (ks *KafkaService) publishEvent(ctx context.Context, eventType string, event interface{}) error {
	// Marshaller l'événement
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erreur marshalling événement: %w", err)
	}

	// Créer le message Kafka
	message := kafka.Message{
		Topic: ks.producerTopic,
		Key:   []byte(eventType),
		Value: eventBytes,
		Headers: []kafka.Header{
			{
				Key:   "event-type",
				Value: []byte(eventType),
			},
			{
				Key:   "timestamp",
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	}

	// Publier le message
	err = ks.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("erreur publication événement: %w", err)
	}

	log.Printf("📤 Événement publié: type=%s topic=%s", eventType, ks.producerTopic)
	return nil
}

// Close ferme les connexions Kafka
func (ks *KafkaService) Close() error {
	var errs []error

	if ks.reader != nil {
		if err := ks.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erreur fermeture reader: %w", err))
		}
	}

	if ks.writer != nil {
		if err := ks.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erreur fermeture writer: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("erreurs fermeture Kafka: %v", errs)
	}

	return nil
}

// VerificarConexion vérifie la connexion à Kafka
func (ks *KafkaService) VerificarConexion(ctx context.Context) error {
	// Essayer de créer une connexion temporaire
	conn, err := kafka.DialContext(ctx, "tcp", ks.bootstrapServers)
	if err != nil {
		return fmt.Errorf("impossible de se connecter à Kafka: %w", err)
	}
	defer conn.Close()

	// Vérifier que le topic existe
	partitions, err := conn.ReadPartitions(ks.topic)
	if err != nil {
		return fmt.Errorf("impossible de lire les partitions du topic %s: %w", ks.topic, err)
	}

	log.Printf("✅ Connexion Kafka vérifiée - Topic: %s, Partitions: %d", ks.topic, len(partitions))
	return nil
}
