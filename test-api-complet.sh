#!/bin/bash

echo "🧪 Tests complets de l'API HistorialBlockchain"
echo "=============================================="

# Configuration
BASE_URL="http://localhost:8081"
API_URL="$BASE_URL/api/historial"

# Couleurs pour les messages
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction pour afficher les résultats
print_result() {
    local status=$1
    local endpoint=$2
    local response=$3
    
    if [ $status -eq 200 ] || [ $status -eq 201 ]; then
        echo -e "${GREEN}✅ SUCCÈS${NC} - $endpoint (HTTP $status)"
    else
        echo -e "${RED}❌ ÉCHEC${NC} - $endpoint (HTTP $status)"
    fi
    
    if [ ! -z "$response" ]; then
        echo -e "${BLUE}📄 Réponse:${NC}"
        echo "$response" | head -10
        echo ""
    fi
}

# Fonction pour faire un appel API
call_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}🔍 Test: $description${NC}"
    echo "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$endpoint")
    elif [ "$method" = "POST" ]; then
        if [ ! -z "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$endpoint")
        else
            response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint")
        fi
    fi
    
    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    print_result $status "$endpoint" "$body"
}

# Démarrage du service
echo "🚀 Démarrage du service..."
cd /home/valentin/Bureau/Cours\ Bogota/Architectura/porojet/BLOCKCHAIN/TEdwin/contenue/historial-blockchain

# Construction
echo "📦 Construction du projet..."
make build > /dev/null 2>&1

# Démarrage en arrière-plan
PORT=8081 ./bin/historial-blockchain > /tmp/service.log 2>&1 &
SERVICE_PID=$!
echo "Service démarré avec le PID: $SERVICE_PID"

# Attendre que le service soit prêt
echo "⏳ Attente du démarrage complet..."
sleep 4

echo -e "\n${BLUE}=== TESTS DES ENDPOINTS DE SANTÉ ===${NC}"

# Test 1: Health Check général
call_api "GET" "$BASE_URL/health" "" "Health Check général"

# Test 2: Readiness Check
call_api "GET" "$BASE_URL/health/ready" "" "Readiness Check"

# Test 3: Liveness Check
call_api "GET" "$BASE_URL/health/live" "" "Liveness Check"

echo -e "\n${BLUE}=== TESTS DES ENDPOINTS D'HISTORIQUE ===${NC}"

# Test 4: Obtenir l'historique d'un produit
call_api "GET" "$API_URL/PROD123" "" "Obtenir historique du produit PROD123"

# Test 5: Obtenir les événements d'un produit avec pagination
call_api "GET" "$API_URL/PROD123/events?page=1&limit=5" "" "Obtenir événements du produit PROD123 (page 1, limit 5)"

# Test 6: Obtenir les événements d'un produit avec filtre de type
call_api "GET" "$API_URL/PROD123/events?tipoEvento=INGRESO&page=1&limit=10" "" "Obtenir événements INGRESO du produit PROD123"

# Test 7: Vérifier un événement spécifique
call_api "GET" "$API_URL/PROD123/verify/EVT456" "" "Vérifier événement EVT456 du produit PROD123"

echo -e "\n${BLUE}=== TESTS DES ENDPOINTS D'ADMINISTRATION ===${NC}"

# Test 8: Reconstruire l'historique
reconstruct_data='{"idProducto":"PROD123","desde":"2024-01-01T00:00:00Z"}'
call_api "POST" "$API_URL/reconstruir" "$reconstruct_data" "Reconstruire historique du produit PROD123"

# Test 9: Obtenir le statut d'une tâche
call_api "GET" "$API_URL/tasks/TASK789" "" "Obtenir statut de la tâche TASK789"

# Test 10: Lister les inconsistances
call_api "GET" "$API_URL/inconsistencies?page=1&limit=10" "" "Lister les inconsistances (page 1, limit 10)"

# Test 11: Lister les inconsistances avec filtre de sévérité
call_api "GET" "$API_URL/inconsistencies?severidad=ALTA&page=1&limit=5" "" "Lister inconsistances de sévérité ALTA"

echo -e "\n${BLUE}=== TESTS DES ENDPOINTS DE MÉTRIQUES ===${NC}"

# Test 12: Métriques
call_api "GET" "$BASE_URL/metrics" "" "Obtenir métriques du service"

echo -e "\n${BLUE}=== TESTS D'ERREURS ET VALIDATION ===${NC}"

# Test 13: Endpoint inexistant
call_api "GET" "$API_URL/inexistant" "" "Test endpoint inexistant (404 attendu)"

# Test 14: Paramètres invalides
call_api "GET" "$API_URL/PROD123/events?page=abc&limit=-1" "" "Test paramètres invalides (400 attendu)"

# Test 15: POST sans données requises
call_api "POST" "$API_URL/reconstruir" '{}' "POST sans données requises (400 attendu)"

# Test 16: ID produit vide
call_api "GET" "$API_URL//events" "" "ID produit vide (400 attendu)"

echo -e "\n${BLUE}=== TESTS DE PERFORMANCE ===${NC}"

# Test 17: Plusieurs appels simultanés
echo -e "\n${YELLOW}🔍 Test: Appels simultanés${NC}"
echo "Lancement de 5 appels simultanés..."

for i in {1..5}; do
    curl -s "$BASE_URL/health" > /tmp/concurrent_$i.txt &
done

wait
echo "Tous les appels simultanés terminés"

# Vérification des résultats
success_count=0
for i in {1..5}; do
    if grep -q "healthy" /tmp/concurrent_$i.txt; then
        success_count=$((success_count + 1))
    fi
done

echo -e "${GREEN}✅ $success_count/5 appels simultanés réussis${NC}"

# Nettoyage des fichiers temporaires
rm -f /tmp/concurrent_*.txt

echo -e "\n${BLUE}=== RÉSUMÉ DES TESTS ===${NC}"
echo "📊 Tests terminés!"
echo "📋 Consultez les logs du service dans /tmp/service.log"

# Arrêt du service
echo -e "\n🛑 Arrêt du service..."
kill $SERVICE_PID 2>/dev/null
wait $SERVICE_PID 2>/dev/null

echo -e "\n${GREEN}✅ Tous les tests sont terminés!${NC}"
echo "📝 Vérifiez les résultats ci-dessus pour voir les détails de chaque test."
