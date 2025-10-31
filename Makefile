# Makefile pour HistorialBlockchain

# Variables
APP_NAME=historial-blockchain
VERSION ?= latest
REGISTRY ?= localhost:5000
GO_VERSION = 1.21

# Couleurs pour les messages
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: help
help: ## Afficher l'aide
	@echo "$(GREEN)📋 Commandes disponibles pour $(APP_NAME):$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: install
install: ## Installer les dépendances
	@echo "$(GREEN)📦 Installation des dépendances...$(NC)"
	go mod download
	go mod tidy

.PHONY: run
run: ## Lancer l'application en mode développement
	@echo "$(GREEN)🚀 Démarrage de l'application...$(NC)"
	go run cmd/api/main.go

.PHONY: build
build: ## Compiler l'application
	@echo "$(GREEN)🔨 Compilation de l'application...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) ./cmd/api

.PHONY: test
test: ## Lancer les tests unitaires
	@echo "$(GREEN)🧪 Exécution des tests unitaires...$(NC)"
	go test -v ./...

.PHONY: test-cover
test-cover: ## Lancer les tests avec couverture
	@echo "$(GREEN)🧪 Exécution des tests avec couverture...$(NC)"
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(YELLOW)📊 Rapport de couverture généré: coverage.html$(NC)"

.PHONY: test-integration
test-integration: ## Lancer les tests d'intégration
	@echo "$(GREEN)🔗 Exécution des tests d'intégration...$(NC)"
	go test -v -tags=integration ./tests/...

.PHONY: lint
lint: ## Lancer le linter
	@echo "$(GREEN)🔍 Analyse du code...$(NC)"
	golangci-lint run

.PHONY: fmt
fmt: ## Formater le code
	@echo "$(GREEN)✨ Formatage du code...$(NC)"
	go fmt ./...

.PHONY: clean
clean: ## Nettoyer les fichiers générés
	@echo "$(GREEN)🧹 Nettoyage...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html

.PHONY: docker-build
docker-build: ## Construire l'image Docker
	@echo "$(GREEN)🐳 Construction de l'image Docker...$(NC)"
	docker build -t $(REGISTRY)/$(APP_NAME):$(VERSION) .

.PHONY: docker-push
docker-push: docker-build ## Pousser l'image Docker vers le registre
	@echo "$(GREEN)📤 Envoi de l'image Docker...$(NC)"
	docker push $(REGISTRY)/$(APP_NAME):$(VERSION)

.PHONY: docker-run
docker-run: ## Lancer l'application avec Docker
	@echo "$(GREEN)🐳 Démarrage avec Docker...$(NC)"
	docker run -p 8081:8081 --env-file .env $(REGISTRY)/$(APP_NAME):$(VERSION)

.PHONY: compose up
compose-up: ## Démarrer tous les services avec Docker Compose
	@echo "$(GREEN)🚀 Démarrage des services avec Docker Compose...$(NC)"
	docker compose up -d

.PHONY: compose-down
compose-down: ## Arrêter tous les services Docker Compose
	@echo "$(GREEN)🛑 Arrêt des services Docker Compose...$(NC)"
	docker compose down

.PHONY: compose-logs
compose-logs: ## Voir les logs Docker Compose
	@echo "$(GREEN)📄 Logs des services...$(NC)"
	docker compose logs -f

.PHONY: dev-setup
dev-setup: ## Configurer l'environnement de développement
	@echo "$(GREEN)🔧 Configuration de l'environnement de développement...$(NC)"
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "$(YELLOW)⚠️  Fichier .env créé. Veuillez le configurer avant de continuer.$(NC)"; \
	fi
	make install

.PHONY: api-test
api-test: ## Tester l'API avec le script de test
	@echo "$(GREEN)🧪 Test de l'API...$(NC)"
	@chmod +x scripts/test-api.sh
	./scripts/test-api.sh

.PHONY: deps-update
deps-update: ## Mettre à jour les dépendances
	@echo "$(GREEN)🔄 Mise à jour des dépendances...$(NC)"
	go get -u ./...
	go mod tidy

.PHONY: security-scan
security-scan: ## Scanner les vulnérabilités de sécurité
	@echo "$(GREEN)🔒 Scan de sécurité...$(NC)"
	gosec ./...

.PHONY: generate-docs
generate-docs: ## Générer la documentation
	@echo "$(GREEN)📚 Génération de la documentation...$(NC)"
	@echo "$(YELLOW)TODO: Implémenter la génération de documentation API$(NC)"

.PHONY: k8s-deploy
k8s-deploy: ## Déployer sur Kubernetes
	@echo "$(GREEN)☸️  Déploiement Kubernetes...$(NC)"
	@if [ -d "k8s" ]; then \
		kubectl apply -f k8s/; \
	else \
		echo "$(RED)❌ Dossier k8s/ non trouvé$(NC)"; \
	fi

.PHONY: k8s-delete
k8s-delete: ## Supprimer le déploiement Kubernetes
	@echo "$(GREEN)🗑️  Suppression du déploiement Kubernetes...$(NC)"
	@if [ -d "k8s" ]; then \
		kubectl delete -f k8s/; \
	else \
		echo "$(RED)❌ Dossier k8s/ non trouvé$(NC)"; \
	fi

.PHONY: load-test
load-test: ## Lancer des tests de charge
	@echo "$(GREEN)⚡ Tests de charge...$(NC)"
	@echo "$(YELLOW)TODO: Implémenter les tests de charge$(NC)"

# Commandes par défaut
.DEFAULT_GOAL := help
