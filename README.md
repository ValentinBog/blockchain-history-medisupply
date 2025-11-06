# Service Blockchain History - MediSupply

## Description
Service de gestion et de reconstruction de l'historique des produits basé sur les événements blockchain. Ce microservice synchronise les données depuis la table `blockchain_medysupply` pour alimenter les tables `historial_transparencia` et `evento_verificado`, permettant ainsi de maintenir un historique cohérent et vérifiable des produits pharmaceutiques.

## Architecture
- **Langage**: Go 1.21+
- **Framework Web**: Gin
- **Base de données**: AWS DynamoDB
- **Message Broker**: Apache Kafka
- **Blockchain**: Ethereum (via RPC)
- **Déploiement**: Docker + Kubernetes

## Tables DynamoDB

### 1. `blockchain_medysupply` (Source principale)
Table contenant tous les événements blockchain des transactions de produits.
```json
{
  "hashEvento": "f28ac63a5723c7f026a37d5cfe951bc4909147b384fab6e44e2d942b0f7db65e",
  "idTransaction": "ef6144d9-35b6-46d9-ab6b-dfc4aaddacc6",
  "idProducto": "PROD-TEST-001",
  "tipoEvento": "fabricacion",
  "actorEmisor": "Laboratorio Medisupply SA",
  "fechaEvento": "2025-11-04T02:10:07.197510032Z",
  "datosEvento": "{\"lote\": \"LOT-12345\", \"fecha_fabricacion\": \"2024-01-15\", \"cantidad\": 1000, \"planta\": \"Planta A\"}",
  "estado": "pendiente",
  "ipfsCid": "QmTqpCZBwK7v8nPFC1Uw8weekn2ZtkbJAQM9VSGC3Lydji",
  "directionBlockchain": "",
  "firmaDigital": "",
  "createdAt": "2025-11-04T02:10:07.227497737Z",
  "updatedAt": "2025-11-04T02:10:07.227497737Z"
}
```

### 2. `historial_transparencia` (Table dérivée)
Historique consolidé par produit.

### 3. `evento_verificado` (Table dérivée)
Événements individuels vérifiés et validés.

## API Endpoints

### 🏥 Endpoints de Santé

#### `GET /health`
**Description**: Vérification basique de l'état du service.

**Réponse**:
```json
{
  "status": "healthy",
  "service": "historial-blockchain",
  "version": "1.0.0"
}
```

#### `GET /health/ready`
**Description**: Vérification de l'état de préparation du service et de ses dépendances.

**Réponse**:
```json
{
  "status": "ready",
  "dependencies": {
    "database": "ok",
    "kafka": "ok", 
    "blockchain": "ok"
  }
}
```

#### `GET /health/live`
**Description**: Vérification de la vivacité du service (liveness probe pour Kubernetes).

**Réponse**:
```json
{
  "status": "alive"
}
```

### 📊 Endpoints Historique

#### `GET /api/historial/{idProducto}`
**Description**: Récupère l'historique complet d'un produit. Le service synchronise automatiquement les données depuis `blockchain_medysupply` avant de retourner l'historique.

**Paramètres**:
- `idProducto` (path): Identifiant unique du produit
- `lote` (query, optionnel): Numéro de lot spécifique
- `full` (query, optionnel): Si `true`, inclut les détails complets des événements

**Exemple de requête**:
```bash
GET /api/historial/PROD-TEST-001?lote=LOT-12345&full=true
```

**Réponse**:
```json
{
  "idProducto": "PROD-TEST-001",
  "lote": "LOT-12345",
  "eventos": [...],
  "estadoActual": "en_transit",
  "fechaCreacion": "2025-11-04T02:10:07Z",
  "ultimaActualizacion": "2025-11-04T12:30:00Z"
}
```

#### `POST /api/historial/reconstruir`
**Description**: Reconstruit l'historique d'un produit à partir des événements blockchain. Supporte le traitement synchrone et asynchrone.

**Paramètres de requête**:
- `async` (query, optionnel): Si `true`, traitement asynchrone

**Corps de la requête**:
```json
{
  "idProducto": "PROD-TEST-001",
  "lote": "LOT-12345", 
  "force": true
}
```

**Réponse synchrone**:
```json
{
  "status": "completed",
  "data": {
    "idProducto": "PROD-TEST-001",
    "eventosReconstruits": 15,
    "inconsistenciasDetectees": 0
  }
}
```

**Réponse asynchrone**:
```json
{
  "status": "processing",
  "taskId": "task-uuid-12345"
}
```

#### `GET /api/historial/{idProducto}/verify/{idEvento}`
**Description**: Vérifie un événement spécifique d'un produit contre la blockchain. Synchronise d'abord les données depuis `blockchain_medysupply`.

**Paramètres**:
- `idProducto` (path): Identifiant du produit
- `idEvento` (path): Identifiant de l'événement

**Exemple**:
```bash
GET /api/historial/PROD-TEST-001/verify/evt-12345
```

**Réponse**:
```json
{
  "idEvento": "evt-12345",
  "idProducto": "PROD-TEST-001",
  "estatVerification": "verified",
  "hashBlockchain": "0x123...",
  "timestampBlockchain": "2025-11-04T02:10:07Z",
  "coherenceDonnees": true,
  "detailsVerification": {
    "blockNumber": 12345678,
    "transactionHash": "0xabc...",
    "gasUsed": 21000
  }
}
```

#### `GET /api/historial/{idProducto}/events`
**Description**: Liste les événements d'un produit avec pagination et filtrage. Utilise directement les données de `blockchain_medysupply`.

**Paramètres**:
- `idProducto` (path): Identifiant du produit
- `tipo` (query, optionnel): Type d'événement à filtrer
- `page` (query, optionnel): Numéro de page (défaut: 1)
- `limit` (query, optionnel): Nombre d'éléments par page (défaut: 10)

**Exemple**:
```bash
GET /api/historial/PROD-TEST-001/events?tipo=fabricacion&page=1&limit=20
```

**Réponse**:
```json
{
  "eventos": [
    {
      "idEvento": "evt-001",
      "tipoEvento": "fabricacion",
      "fechaEvento": "2025-11-04T02:10:07Z",
      "actorEmisor": "Laboratorio Medisupply SA",
      "estado": "confirmado",
      "datosEvento": {...}
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 15
  }
}
```

#### `GET /api/historial/tasks/{taskId}`
**Description**: Récupère le statut d'une tâche de reconstruction asynchrone.

**Paramètres**:
- `taskId` (path): Identifiant de la tâche

**Exemple**:
```bash
GET /api/historial/tasks/task-uuid-12345
```

**Réponse**:
```json
{
  "taskId": "task-uuid-12345",
  "status": "completed",
  "progress": 100,
  "startTime": "2025-11-04T02:10:07Z",
  "endTime": "2025-11-04T02:15:30Z",
  "result": {
    "eventosTraites": 15,
    "inconsistenciasDetectees": 0,
    "erreursRencontrees": 0
  },
  "error": null
}
```

#### `GET /api/historial/inconsistencies`
**Description**: Liste les inconsistances détectées dans les historiques avec pagination et filtrage par sévérité.

**Paramètres**:
- `severidad` (query, optionnel): Filtre par sévérité (critique, majeure, mineure)
- `page` (query, optionnel): Numéro de page (défaut: 1)  
- `limit` (query, optionnel): Éléments par page (défaut: 50)

**Exemple**:
```bash
GET /api/historial/inconsistencies?severidad=critique&page=1&limit=25
```

**Réponse**:
```json
{
  "inconsistencias": [
    {
      "id": "inc-001",
      "idProducto": "PROD-TEST-001", 
      "tipoInconsistencia": "hash_mismatch",
      "severidad": "critique",
      "description": "Hash de l'événement ne correspond pas à la blockchain",
      "fechaDeteccion": "2025-11-04T10:30:00Z",
      "resuelto": false
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 25,
    "total": 3
  },
  "filtres": {
    "severidad": "critique"
  }
}
```

### 📈 Endpoint Métriques

#### `GET /metrics`
**Description**: Endpoint pour les métriques Prometheus (si activé dans la configuration).

**Réponse**: Format Prometheus metrics

## Synchronisation des Données

Le service utilise une stratégie de synchronisation intelligente :

1. **Synchronisation automatique**: Avant chaque opération de lecture (`ObtenerHistorial`, `VerificarEvento`), le service synchronise les données depuis `blockchain_medysupply`

2. **Synchronisation en temps réel**: Via le consumer Kafka qui écoute les nouveaux événements

3. **Reconstruction complète**: Via l'endpoint `/reconstruir` pour forcer une reconstruction complète

### Flux de Synchronisation

```
blockchain_medysupply (source)
        ↓
    Synchronisation
        ↓
evento_verificado + historial_transparencia (dérivées)
```

## Configuration

Variables d'environnement principales :

```env
# DynamoDB
DYNAMODB_TABLE_HISTORIAL=historial_transparencia
DYNAMODB_TABLE_EVENTO=evento_verificado  
DYNAMODB_TABLE_BLOCKCHAIN_EVENTS=blockchain_medysupply

# Kafka
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_TOPIC=event.transaccion.blockchain.registered

# Blockchain
BLOCKCHAIN_RPC_URL=http://localhost:8545
ENABLE_STRICT_VERIFICATION=false

# Serveur
SERVER_PORT=8081
```

## Démarrage Rapide

### Avec Docker
```bash
docker-compose up -d
```

### Manuel
```bash
# Installation des dépendances
go mod download

# Compilation
make build

# Exécution
make run
```

## Tests

```bash
# Tests unitaires
make test

# Tests d'intégration
make test-integration

# Tests API complets
./test-api-complete.sh
```

## Monitoring et Observabilité

- **Logs**: Structurés en JSON avec niveaux de log configurables
- **Métriques**: Support Prometheus (si activé)
- **Tracing**: Support Jaeger (si configuré)
- **Health Checks**: Endpoints dédiés pour Kubernetes

## Codes de Statut HTTP

- `200 OK`: Succès
- `201 Created`: Ressource créée
- `202 Accepted`: Traitement asynchrone accepté
- `400 Bad Request`: Paramètres invalides
- `404 Not Found`: Ressource non trouvée
- `429 Too Many Requests`: Limite de débit dépassée
- `500 Internal Server Error`: Erreur serveur

## Support et Contribution

Pour toute question ou contribution, veuillez consulter la documentation technique dans le dossier `/docs`.