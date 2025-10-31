#!/bin/bash

echo "🧪 Test API HistorialBlockchain avec données réelles"
echo "=================================================="

BASE_URL="http://localhost:8081"

# Variables d'environnement pour DynamoDB local
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test  
export AWS_REGION=us-east-1
export DYNAMODB_ENDPOINT=http://localhost:8000

echo "🚀 Démarrage du service..."
cd /home/valentin/Bureau/Cours\ Bogota/Architectura/porojet/BLOCKCHAIN/TEdwin/contenue/historial-blockchain

# Démarrage avec les bonnes variables d'environnement
PORT=8081 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 DYNAMODB_ENDPOINT=http://localhost:8000 ./bin/historial-blockchain &
SERVICE_PID=$!

echo "Service démarré avec PID: $SERVICE_PID"
sleep 3

echo ""
echo "🔍 Tests des nouveaux endpoints implémentés:"
echo "=============================================="

echo ""
echo "1. Health check:"
curl -s "$BASE_URL/health" | head -5
echo -e "\n"

echo "2. Obtenir événements du produit PROD123:"
curl -s "$BASE_URL/api/historial/PROD123/events?page=1&limit=5" | head -10
echo -e "\n"

echo "3. Obtenir événements de type INGRESO:"
curl -s "$BASE_URL/api/historial/PROD123/events?tipo=INGRESO&page=1&limit=5" | head -10
echo -e "\n"

echo "4. Lister les inconsistances:"
curl -s "$BASE_URL/api/historial/inconsistencies?page=1&limit=5" | head -10
echo -e "\n"

echo "5. Lister inconsistances de sévérité ALTA:"
curl -s "$BASE_URL/api/historial/inconsistencies?severidad=ALTA&page=1&limit=5" | head -10
echo -e "\n"

echo "🛑 Arrêt du service..."
kill $SERVICE_PID 2>/dev/null
wait $SERVICE_PID 2>/dev/null

echo "✅ Tests terminés!"
