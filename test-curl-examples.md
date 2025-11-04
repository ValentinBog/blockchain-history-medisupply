# Exemples de requêtes curl pour tester l'API
# blockchain-history-medisupply

# Base URL
BASE_URL="http://localhost:8081"

echo "🧪 Tests API blockchain-history-medisupply"
echo "==========================================="
echo ""

# 1. Test de santé
echo "1️⃣ Test de santé du service"
echo "curl -X GET $BASE_URL/health"
echo ""

# 2. Test de readiness
echo "2️⃣ Test de disponibilité"
echo "curl -X GET $BASE_URL/health/ready"
echo ""

# 3. Test de liveness
echo "3️⃣ Test de vivacité"
echo "curl -X GET $BASE_URL/health/live"
echo ""

# 4. Obtenir un historique (simulé)
echo "4️⃣ Obtenir l'historique d'un produit"
echo "curl -X GET $BASE_URL/api/historial/PROD123"
echo ""

# 5. Obtenir un historique avec lote
echo "5️⃣ Obtenir l'historique d'un produit avec lote"
echo "curl -X GET '$BASE_URL/api/historial/PROD123?lote=L001'"
echo ""

# 6. Obtenir un historique complet
echo "6️⃣ Obtenir l'historique complet"
echo "curl -X GET '$BASE_URL/api/historial/PROD123?full=true'"
echo ""

# 7. Reconstruire un historique (synchrone)
echo "7️⃣ Reconstruire un historique (mode synchrone)"
echo 'curl -X POST $BASE_URL/api/historial/reconstruir \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '"'"'{'
echo '    "idProducto": "PROD123",'
echo '    "lote": "L001",'
echo '    "force": true'
echo '  }'"'"
echo ""

# 8. Reconstruire un historique (asynchrone)
echo "8️⃣ Reconstruire un historique (mode asynchrone)"
echo 'curl -X POST "$BASE_URL/api/historial/reconstruir?async=true" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '"'"'{'
echo '    "idProducto": "PROD456",'
echo '    "lote": "L002",'
echo '    "force": false'
echo '  }'"'"
echo ""

# 9. Vérifier un événement spécifique
echo "9️⃣ Vérifier un événement spécifique"
echo "curl -X GET $BASE_URL/api/historial/PROD123/verify/EVT456"
echo ""

# 10. Obtenir les événements d'un produit
echo "🔟 Obtenir les événements d'un produit (paginé)"
echo "curl -X GET '$BASE_URL/api/historial/PROD123/events?page=1&limit=5'"
echo ""

# 11. Filtrer les événements par type
echo "1️⃣1️⃣ Filtrer les événements par type"
echo "curl -X GET '$BASE_URL/api/historial/PROD123/events?tipo=INGRESO&page=1&limit=10'"
echo ""

# 12. Obtenir le statut d'une tâche
echo "1️⃣2️⃣ Obtenir le statut d'une tâche asynchrone"
echo "curl -X GET $BASE_URL/api/historial/tasks/TASK-UUID-HERE"
echo ""

# 13. Lister les inconsistances
echo "1️⃣3️⃣ Lister toutes les inconsistances"
echo "curl -X GET '$BASE_URL/api/historial/inconsistencies?page=1&limit=20'"
echo ""

# 14. Filtrer les inconsistances par sévérité
echo "1️⃣4️⃣ Filtrer les inconsistances par sévérité"
echo "curl -X GET '$BASE_URL/api/historial/inconsistencies?severidad=ALTA&page=1&limit=10'"
echo ""

# 15. Test avec jq pour un JSON formaté
echo "1️⃣5️⃣ Test avec JSON formaté (nécessite jq)"
echo "curl -s -X GET $BASE_URL/health | jq ."
echo ""

echo "💡 Notes:"
echo "- Remplacez TASK-UUID-HERE par un vrai UUID de tâche"
echo "- Installez jq pour un JSON formaté: sudo apt install jq"
echo "- Les données sont simulées dans le service pour les tests"
