# 🧪 Guide de Test - blockchain-history-medisupply

## 📋 Prérequis

- Docker et Docker Compose installés
- Go 1.19+ installé
- AWS CLI installé (pour créer les tables DynamoDB)
- curl installé pour les tests
- (Optionnel) jq pour formatter le JSON

## 🚀 Étapes de Configuration et Test

### 1️⃣ Démarrer les Services de Dépendance

```bash
# Démarrer DynamoDB local et Kafka
docker-compose -f docker-compose.test.yml up -d

# Vérifier que les services sont démarrés
docker-compose -f docker-compose.test.yml ps
```

### 2️⃣ Initialiser les Tables DynamoDB

```bash
# Attendre que DynamoDB soit prêt (environ 10-15 secondes)
sleep 15

# Créer les tables
./init-tables.sh
```

### 3️⃣ Compiler et Démarrer le Service

```bash
# Compiler le service
make build

# OU manuellement
go build -o bin/historial-blockchain cmd/api/main.go

# Démarrer le service
./bin/historial-blockchain
```

### 4️⃣ Tester avec curl

Le service démarre sur le port **8081**. Voir `test-curl-examples.md` pour tous les exemples.

**Tests de base :**

```bash
# Test de santé
curl -X GET http://localhost:8081/health

# Test d'historique (données simulées)
curl -X GET http://localhost:8081/api/historial/PROD123

# Test de reconstruction
curl -X POST http://localhost:8081/api/historial/reconstruir \
  -H "Content-Type: application/json" \
  -d '{"idProducto":"PROD123","lote":"L001","force":true}'
```

## 🔧 Interfaces Web Disponibles

- **DynamoDB Admin** : http://localhost:8001
- **Kafka UI** : http://localhost:8090

## 📊 Vérification des Tables

```bash
# Lister les tables DynamoDB
aws dynamodb list-tables --endpoint-url http://localhost:8000 --region us-east-1

# Voir le contenu d'une table
aws dynamodb scan --table-name historial_transparencia --endpoint-url http://localhost:8000 --region us-east-1
```

## 🛑 Arrêter les Services

```bash
# Arrêter le service Go (Ctrl+C dans le terminal)

# Arrêter les conteneurs Docker
docker-compose -f docker-compose.test.yml down

# Nettoyer les volumes (optionnel)
docker-compose -f docker-compose.test.yml down -v
```

## 🐛 Dépannage

### Problème : "Cannot connect to DynamoDB"
- Vérifier que le conteneur DynamoDB est démarré : `docker ps`
- Vérifier l'endpoint dans `.env` : `DYNAMODB_ENDPOINT=http://localhost:8000`

### Problème : "Cannot connect to Kafka"
- Vérifier que Kafka est démarré : `docker logs kafka-test`
- Le service peut démarrer sans Kafka (mode dégradé)

### Problème : "Table does not exist"
- Relancer `./init-tables.sh`
- Vérifier les tables : `aws dynamodb list-tables --endpoint-url http://localhost:8000 --region us-east-1`

## 📝 Notes Importantes

1. **Mode Test** : La vérification blockchain stricte est désactivée (données simulées)
2. **Données Simulées** : Les réponses des endpoints utilisent des données de test
3. **Persistence** : Les données DynamoDB sont persistées dans des volumes Docker
4. **Rate Limiting** : Désactivé pour faciliter les tests

## 🎯 Endpoints Principaux à Tester

| Endpoint | Méthode | Description |
|----------|---------|-------------|
| `/health` | GET | Santé du service |
| `/api/historial/{id}` | GET | Obtenir historique |
| `/api/historial/reconstruir` | POST | Reconstruire historique |
| `/api/historial/{id}/verify/{event}` | GET | Vérifier événement |
| `/api/historial/{id}/events` | GET | Lister événements |
| `/api/historial/inconsistencies` | GET | Lister inconsistances |

Consultez `test-curl-examples.md` pour tous les exemples de requêtes.
