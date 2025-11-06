package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config representa la configuración de la aplicación HistorialBlockchain
type Config struct {
	// AWS
	AWSRegion         string
	AWSAccessKeyID    string
	AWSSecretKey      string
	DynamoDBTableHistorial string
	DynamoDBTableEvento    string
	DynamoDBTableBlockchainEvents string
	DynamoDBEndpoint       string
	UseAWSSecrets     bool

	// Kafka
	KafkaBootstrapServers string
	KafkaConsumerGroup    string
	KafkaTopic           string
	KafkaProducerTopic   string

	// Blockchain
	AlchemyAPIKey     string
	BlockchainRPCURL  string
	BlockchainNetwork string

	// Server
	ServerPort string
	GinMode    string

	// Security
	EncryptionKey string
	JWTPublicKey  string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int

	// Observabilidad
	EnableTracing   bool
	JaegerEndpoint  string
	EnableMetrics   bool
	PrometheusPort  string

	// Blockchain Verification
	EnableStrictVerification bool
	BlockchainTimeout       int
	MaxRetries             int
}

var AppConfig *Config

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() (*Config, error) {
	// Intentar cargar .env solo en desarrollo
	_ = godotenv.Load()

	config := &Config{
		// AWS
		AWSRegion:              getEnvOrDefault("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:         os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:          os.Getenv("AWS_SECRET_ACCESS_KEY"),
		DynamoDBTableHistorial: getEnvOrDefault("DYNAMODB_TABLE_HISTORIAL", "historial_transparencia"),
		DynamoDBTableEvento:    getEnvOrDefault("DYNAMODB_TABLE_EVENTO", "evento_verificado"),
		DynamoDBTableBlockchainEvents: getEnvOrDefault("DYNAMODB_TABLE_BLOCKCHAIN_EVENTS", "blockcahin_medysupyly"),
		DynamoDBEndpoint:       os.Getenv("DYNAMODB_ENDPOINT"),
		UseAWSSecrets:         getEnvAsBool("USE_AWS_SECRETS", false),

		// Kafka
		KafkaBootstrapServers: getEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaConsumerGroup:    getEnvOrDefault("KAFKA_CONSUMER_GROUP", "historial-blockchain-consumer"),
		KafkaTopic:           getEnvOrDefault("KAFKA_TOPIC", "event.transaccion.blockchain.registered"),
		KafkaProducerTopic:   getEnvOrDefault("KAFKA_PRODUCER_TOPIC", "event.historial"),

		// Blockchain
		AlchemyAPIKey:     os.Getenv("ALCHEMY_API_KEY"),
		BlockchainRPCURL:  getEnvOrDefault("BLOCKCHAIN_RPC_URL", ""),
		BlockchainNetwork: getEnvOrDefault("BLOCKCHAIN_NETWORK", "sepolia"),

		// Server
		ServerPort: getEnvOrDefault("SERVER_PORT", "8081"),
		GinMode:    getEnvOrDefault("GIN_MODE", "debug"),

		// Security
		EncryptionKey: getEnvOrDefault("ENCRYPTION_KEY", "changeme-32-chars-encryption-key"),
		JWTPublicKey:  os.Getenv("JWT_PUBLIC_KEY"),

		// Rate Limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsInt("RATE_LIMIT_WINDOW", 60),

		// Observabilidad
		EnableTracing:  getEnvAsBool("ENABLE_TRACING", true),
		JaegerEndpoint: getEnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		EnableMetrics:  getEnvAsBool("ENABLE_METRICS", true),
		PrometheusPort: getEnvOrDefault("PROMETHEUS_PORT", "2112"),

		// Blockchain Verification
		EnableStrictVerification: getEnvAsBool("ENABLE_STRICT_VERIFICATION", true),
		BlockchainTimeout:       getEnvAsInt("BLOCKCHAIN_TIMEOUT", 30),
		MaxRetries:             getEnvAsInt("MAX_RETRIES", 3),
	}

	// Construir URL de blockchain si no se proporciona
	if config.BlockchainRPCURL == "" && config.AlchemyAPIKey != "" {
		config.BlockchainRPCURL = fmt.Sprintf("https://eth-%s.g.alchemy.com/v2/%s",
			config.BlockchainNetwork, config.AlchemyAPIKey)
	}

	// Validar configuración crítica
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	AppConfig = config
	return config, nil
}

func validateConfig(config *Config) error {
	if config.KafkaBootstrapServers == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS es requerido")
	}

	if config.BlockchainRPCURL == "" {
		return fmt.Errorf("BLOCKCHAIN_RPC_URL o ALCHEMY_API_KEY es requerido")
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
