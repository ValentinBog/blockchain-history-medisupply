#!/bin/bash

echo "🔧 Initialisation de l'environnement de test"
echo "============================================"

# Configuration
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
export DYNAMODB_ENDPOINT=http://localhost:8000

# Démarrage de DynamoDB local
echo "🚀 Démarrage de DynamoDB local..."
docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest > /dev/null 2>&1

# Attendre que DynamoDB soit prêt
echo "⏳ Attente que DynamoDB soit prêt..."
sleep 5

# Vérification de DynamoDB
echo "🔍 Vérification de DynamoDB..."
curl -s http://localhost:8000 > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ DynamoDB local est démarré"
else
    echo "❌ Erreur: DynamoDB local ne répond pas"
    exit 1
fi

# Création des tables DynamoDB
echo "📊 Création des tables DynamoDB..."

# Table pour l'historique des produits
aws dynamodb create-table \
    --endpoint-url http://localhost:8000 \
    --table-name HistorialProductos \
    --attribute-definitions \
        AttributeName=IDProducto,AttributeType=S \
    --key-schema \
        AttributeName=IDProducto,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --no-cli-pager > /dev/null 2>&1

# Table pour les événements
aws dynamodb create-table \
    --endpoint-url http://localhost:8000 \
    --table-name EventosHistorial \
    --attribute-definitions \
        AttributeName=IDEvento,AttributeType=S \
        AttributeName=IDProducto,AttributeType=S \
    --key-schema \
        AttributeName=IDEvento,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=ProductoIndex,KeySchema=[{AttributeName=IDProducto,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --no-cli-pager > /dev/null 2>&1

# Table pour les tâches
aws dynamodb create-table \
    --endpoint-url http://localhost:8000 \
    --table-name TareasReconstruccion \
    --attribute-definitions \
        AttributeName=TaskID,AttributeType=S \
    --key-schema \
        AttributeName=TaskID,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --no-cli-pager > /dev/null 2>&1

echo "✅ Tables DynamoDB créées"

# Insertion de données de test
echo "📝 Insertion de données de test..."

# Historique de produit de test
aws dynamodb put-item \
    --endpoint-url http://localhost:8000 \
    --table-name HistorialProductos \
    --item '{
        "IDProducto": {"S": "PROD123"},
        "FechaCreacion": {"S": "2024-01-15T10:30:00Z"},
        "FechaActualizacion": {"S": "2024-10-31T15:00:00Z"},
        "TotalEventos": {"N": "3"},
        "EstadoActual": {"S": "EN_STOCK"},
        "UltimaVerificacion": {"S": "2024-10-31T14:00:00Z"},
        "EventosIds": {"SS": ["EVT456", "EVT789", "EVT101"]}
    }' \
    --no-cli-pager > /dev/null 2>&1

# Eventos de test
aws dynamodb put-item \
    --endpoint-url http://localhost:8000 \
    --table-name EventosHistorial \
    --item '{
        "IDEvento": {"S": "EVT456"},
        "IDProducto": {"S": "PROD123"},
        "TipoEvento": {"S": "INGRESO"},
        "Fecha": {"S": "2024-01-15T10:30:00Z"},
        "ReferenciaBlockchain": {"S": "0x123abc..."},
        "EstadoVerificacion": {"S": "VERIFICADO"},
        "Detalles": {"M": {
            "cantidad": {"N": "100"},
            "lote": {"S": "L001"},
            "proveedor": {"S": "PROV001"}
        }}
    }' \
    --no-cli-pager > /dev/null 2>&1

aws dynamodb put-item \
    --endpoint-url http://localhost:8000 \
    --table-name EventosHistorial \
    --item '{
        "IDEvento": {"S": "EVT789"},
        "IDProducto": {"S": "PROD123"},
        "TipoEvento": {"S": "EGRESO"},
        "Fecha": {"S": "2024-02-15T14:20:00Z"},
        "ReferenciaBlockchain": {"S": "0x456def..."},
        "EstadoVerificacion": {"S": "VERIFICADO"},
        "Detalles": {"M": {
            "cantidad": {"N": "50"},
            "destino": {"S": "HOSPITAL_001"}
        }}
    }' \
    --no-cli-pager > /dev/null 2>&1

# Tâche de test
aws dynamodb put-item \
    --endpoint-url http://localhost:8000 \
    --table-name TareasReconstruccion \
    --item '{
        "TaskID": {"S": "TASK789"},
        "IDProducto": {"S": "PROD123"},
        "Estado": {"S": "COMPLETADA"},
        "FechaInicio": {"S": "2024-10-30T10:00:00Z"},
        "FechaFin": {"S": "2024-10-30T10:05:00Z"},
        "Progreso": {"N": "100"},
        "EventosProcesados": {"N": "3"},
        "Errores": {"N": "0"}
    }' \
    --no-cli-pager > /dev/null 2>&1

echo "✅ Données de test insérées"

echo ""
echo "🎯 Environnement de test prêt!"
echo "📊 Tables créées: HistorialProductos, EventosHistorial, TareasReconstruccion"
echo "📝 Données de test disponibles pour PROD123"
echo ""
echo "🚀 Vous pouvez maintenant lancer les tests de l'API:"
echo "   ./test-api-complet.sh"
echo ""
echo "🛑 Pour arrêter l'environnement:"
echo "   docker stop dynamodb-local && docker rm dynamodb-local"
