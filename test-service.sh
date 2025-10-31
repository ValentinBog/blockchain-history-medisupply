#!/bin/bash

echo "🧪 Test du microservice HistorialBlockchain"
echo "=============================================="

cd /home/valentin/Bureau/Cours\ Bogota/Architectura/porojet/BLOCKCHAIN/TEdwin/contenue/historial-blockchain

echo "📦 Construction du projet..."
make build

if [ $? -ne 0 ]; then
    echo "❌ Erreur de construction"
    exit 1
fi

echo "🚀 Démarrage du service sur le port 8081..."
PORT=8081 ./bin/historial-blockchain &
SERVICE_PID=$!
echo "Service démarré avec le PID: $SERVICE_PID"

# Attendre que le service soit prêt
echo "⏳ Attente du démarrage complet..."
sleep 3

# Test des endpoints
echo "🩺 Test de l'endpoint de santé..."
curl -s -w "\nCode de statut: %{http_code}\n" http://localhost:8081/health

echo -e "\n🔍 Test de l'endpoint de préparation..."
curl -s -w "\nCode de statut: %{http_code}\n" http://localhost:8081/health/ready

echo -e "\n📊 Test d'un endpoint de l'API..."
curl -s -w "\nCode de statut: %{http_code}\n" "http://localhost:8081/api/historial/PROD123/events?page=1&limit=10"

echo -e "\n🛑 Arrêt du service..."
kill $SERVICE_PID
wait $SERVICE_PID 2>/dev/null

echo "✅ Tests terminés!"
