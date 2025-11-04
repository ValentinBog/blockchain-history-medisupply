#!/bin/bash

# Script de test pour l'API blockchain-history-medisupply
# Tests des endpoints principaux

BASE_URL="http://localhost:8081"
CONTENT_TYPE="Content-Type: application/json"

echo "🧪 Tests de l'API blockchain-history-medisupply"
echo "================================================"

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour tester un endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}🔍 Test: ${description}${NC}"
    echo "   ${method} ${endpoint}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "${BASE_URL}${endpoint}")
    else
        response=$(curl -s -w "\n%{http_code}" -X "${method}" -H "${CONTENT_TYPE}" -d "${data}" "${BASE_URL}${endpoint}")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "   ${GREEN}✅ Succès (${http_code})${NC}"
        echo "   Réponse: $(echo "$body" | jq -C . 2>/dev/null || echo "$body")"
    else
        echo -e "   ${RED}❌ Échec (${http_code})${NC}"
        echo "   Erreur: $body"
    fi
}

# Vérifier si le service est démarré
echo "🔍 Vérification de la disponibilité du service..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo -e "${RED}❌ Service non disponible sur ${BASE_URL}${NC}"
    echo "   Assurez-vous que le service est démarré avec 'make compose-up'"
    exit 1
fi

echo -e "${GREEN}✅ Service disponible${NC}"

# Tests des endpoints de santé
echo -e "\n${YELLOW}📋 Tests des endpoints de santé${NC}"
test_endpoint "GET" "/health" "" "Health check basique"
test_endpoint "GET" "/health/ready" "" "Readiness check"
test_endpoint "GET" "/health/live" "" "Liveness check"

# Tests des endpoints d'historique
echo -e "\n${YELLOW}📋 Tests des endpoints d'historique${NC}"

# Test récupération historique
test_endpoint "GET" "/api/historial/PROD123" "" "Récupérer historique produit PROD123"
test_endpoint "GET" "/api/historial/PROD123?lote=L001" "" "Récupérer historique produit PROD123 lote L001"
test_endpoint "GET" "/api/historial/PROD123?full=true" "" "Récupérer historique complet produit PROD123"

# Test reconstruction historique (synchrone)
reconstruction_data='{
  "idProducto": "PROD123",
  "lote": "L001",
  "force": true
}'
test_endpoint "POST" "/api/historial/reconstruir" "$reconstruction_data" "Reconstruction synchrone historique"

# Test reconstruction historique (asynchrone)
test_endpoint "POST" "/api/historial/reconstruir?async=true" "$reconstruction_data" "Reconstruction asynchrone historique"

# Test vérification d'événement
test_endpoint "GET" "/api/historial/PROD123/verify/EVT456" "" "Vérifier événement EVT456"

# Test récupération d'événements
test_endpoint "GET" "/api/historial/PROD123/events" "" "Récupérer événements produit PROD123"
test_endpoint "GET" "/api/historial/PROD123/events?tipo=INGRESO" "" "Récupérer événements INGRESO"
test_endpoint "GET" "/api/historial/PROD123/events?page=1&limit=5" "" "Récupérer événements avec pagination"

# Test récupération des inconsistances
test_endpoint "GET" "/api/historial/inconsistencies" "" "Lister toutes les inconsistances"
test_endpoint "GET" "/api/historial/inconsistencies?severidad=ALTA" "" "Lister inconsistances sévérité ALTA"
test_endpoint "GET" "/api/historial/inconsistencies?page=1&limit=10" "" "Lister inconsistances avec pagination"

# Tests des cas d'erreur
echo -e "\n${YELLOW}📋 Tests des cas d'erreur${NC}"
test_endpoint "GET" "/api/historial/" "" "Historique sans ID produit (erreur attendue)"
test_endpoint "GET" "/api/historial/INEXISTANT" "" "Historique produit inexistant"
test_endpoint "GET" "/api/historial/PROD123/verify/" "" "Vérification sans ID événement (erreur attendue)"

# Test données invalides
invalid_data='{"invalid": "data"}'
test_endpoint "POST" "/api/historial/reconstruir" "$invalid_data" "Reconstruction avec données invalides (erreur attendue)"

echo -e "\n${GREEN}🎉 Tests terminés!${NC}"
echo ""
echo "📊 Pour voir les métriques:"
echo "   - Kafka UI: http://localhost:8090"
echo "   - DynamoDB Admin: http://localhost:8001"
echo ""
