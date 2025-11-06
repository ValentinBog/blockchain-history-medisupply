package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"

	appConfig "github.com/edinfamous/historial-blockchain/internal/config"
	"github.com/edinfamous/historial-blockchain/internal/handlers"
	"github.com/edinfamous/historial-blockchain/internal/middleware"
	"github.com/edinfamous/historial-blockchain/internal/models"
	"github.com/edinfamous/historial-blockchain/internal/services"
)

func main() {
	// Charger la configuration
	cfg, err := appConfig.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Erreur chargement configuration: %v", err)
	}

	log.Println("✅ Configuration chargée correctement")

	// Configurer Gin
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialiser les services
	log.Println("🔧 Initialisation des services...")

	// 1. Initialiser DynamoDB
	dynamoClient, err := initDynamoDBClient(cfg)
	if err != nil {
		log.Fatalf("❌ Erreur initialisation DynamoDB: %v", err)
	}
	log.Println("✅ Connecté à DynamoDB")

	dynamoDBService := services.NewDynamoDBService(
		dynamoClient,
		cfg.DynamoDBTableHistorial,
		cfg.DynamoDBTableEvento,
		cfg.DynamoDBTableBlockchainEvents,
	)

	// 2. Initialiser Blockchain Service
	blockchainService, err := services.NewBlockchainService(
		cfg.BlockchainRPCURL,
		time.Duration(cfg.BlockchainTimeout)*time.Second,
		cfg.MaxRetries,
	)
	if err != nil {
		log.Printf("⚠️ Avertissement: Impossible de se connecter à la blockchain: %v", err)
		log.Println("   Le service continuera sans vérification blockchain stricte")
		cfg.EnableStrictVerification = false
	} else {
		err = blockchainService.VerificarConexion(context.Background())
		if err != nil {
			log.Printf("⚠️ Avertissement: Connexion blockchain échouée: %v", err)
			cfg.EnableStrictVerification = false
		} else {
			log.Println("✅ Connecté à la blockchain")
		}
	}

	// 3. Initialiser Kafka Service
	kafkaService := services.NewKafkaService(
		cfg.KafkaBootstrapServers,
		cfg.KafkaConsumerGroup,
		cfg.KafkaTopic,
		cfg.KafkaProducerTopic,
	)

	err = kafkaService.VerificarConexion(context.Background())
	if err != nil {
		log.Printf("⚠️ Avertissement: Connexion Kafka échouée: %v", err)
		log.Println("   Assurez-vous que Kafka est en cours d'exécution")
	} else {
		log.Println("✅ Connecté à Kafka")
	}

	// 4. Initialiser Historial Service
	historialService := services.NewHistorialService(
		dynamoDBService,
		blockchainService,
		kafkaService,
		cfg.EnableStrictVerification,
	)

	// Initialiser les handlers
	healthHandler := handlers.NewHealthHandler()
	historialHandler := handlers.NewHistorialHandler(historialService)

	// Configurer les routes
	router := setupRoutes(cfg, healthHandler, historialHandler)

	// Créer le serveur HTTP
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Démarrer le consumer Kafka en arrière-plan
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("🎧 Démarrage du consumer Kafka...")
		
		// Wrapper pour adapter la signature de la fonction
		handler := func(event *models.TransaccionBlockchainEvent) error {
			return historialService.TraiterEvenementTransaccion(ctx, event)
		}
		
		err := kafkaService.ConsumeEvents(ctx, handler)
		if err != nil && ctx.Err() == nil {
			log.Printf("❌ Erreur consumer Kafka: %v", err)
		}
	}()

	// Démarrer le serveur HTTP
	go func() {
		log.Printf("🚀 Serveur démarré sur le port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erreur serveur: %v", err)
		}
	}()

	// Attendre le signal d'arrêt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Arrêt du serveur...")

	// Arrêter gracieusement
	cancel() // Arrêter le consumer Kafka

	// Arrêter le serveur HTTP avec timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Erreur arrêt serveur: %v", err)
	}

	// Attendre que le consumer se termine
	wg.Wait()

	// Fermer les services
	if kafkaService != nil {
		kafkaService.Close()
	}
	if blockchainService != nil {
		blockchainService.Close()
	}

	log.Println("✅ Serveur arrêté proprement")
}

// initDynamoDBClient initialise le client DynamoDB
func initDynamoDBClient(cfg *appConfig.Config) (*dynamodb.Client, error) {
	ctx := context.Background()

	var awsConfig aws.Config
	var err error

	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretKey != "" {
		// Utiliser les credentials fournis
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.AWSRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKeyID,
				cfg.AWSSecretKey,
				"",
			)),
		)
	} else {
		// Utiliser les credentials par défaut (IAM role, env vars, etc.)
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.AWSRegion),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("impossible de charger la configuration AWS: %w", err)
	}

	// Créer le client DynamoDB
	dynamoClient := dynamodb.NewFromConfig(awsConfig)

	// Si un endpoint local est configuré, l'utiliser
	if cfg.DynamoDBEndpoint != "" {
		dynamoClient = dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.DynamoDBEndpoint)
		})
		log.Printf("🔧 Utilisation de l'endpoint DynamoDB local: %s", cfg.DynamoDBEndpoint)
	}

	return dynamoClient, nil
}

// setupRoutes configure les routes de l'application
func setupRoutes(cfg *appConfig.Config, healthHandler *handlers.HealthHandler, historialHandler *handlers.HistorialHandler) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(middleware.SetupLogging())
	router.Use(gin.Recovery())
	router.Use(middleware.SetupCORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.CorrelationID())

	// Rate limiting si configuré
	if cfg.RateLimitRequests > 0 {
		rateLimitConfig := middleware.RateLimitConfig{
			RequestsPerSecond: cfg.RateLimitRequests / cfg.RateLimitWindow,
			BurstSize:         cfg.RateLimitRequests,
		}
		router.Use(middleware.SetupRateLimit(rateLimitConfig))
	}

	// Routes de santé
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/ready", healthHandler.ReadinessCheck)
	router.GET("/health/live", healthHandler.LivenessCheck)

	// Groupe API
	apiGroup := router.Group("/api")
	{
		// Routes historial
		historialGroup := apiGroup.Group("/historial")
		{
			historialGroup.GET("/:idProducto", historialHandler.ObtenerHistorial)
			historialGroup.POST("/reconstruir", historialHandler.ReconstruirHistorial)
			historialGroup.GET("/:idProducto/verify/:idEvento", historialHandler.VerificarEvento)
			historialGroup.GET("/:idProducto/events", historialHandler.ObtenerEventos)
			historialGroup.GET("/tasks/:taskId", historialHandler.ObtenerStatusTarea)
			historialGroup.GET("/inconsistencies", historialHandler.ListarInconsistencias)
		}
	}

	// Route pour metrics Prometheus (si activé)
	if cfg.EnableMetrics {
		// router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		// Pour l'instant, juste un placeholder
		router.GET("/metrics", func(c *gin.Context) {
			c.String(http.StatusOK, "# Metrics endpoint - TODO: implement Prometheus metrics")
		})
	}

	return router
}
