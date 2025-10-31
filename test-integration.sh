#!/bin/bash

echo "🔗 Test d'intégration entre TransaccionBlockchain et HistorialBlockchain"
echo "======================================================================="

# URLs des services
TRANSACCION_API="http://localhost:8080/api/transacciones"
HISTORIAL_API="http://localhost:8081/api/historial"

echo "🧪 Étape 1: Création d'événements via TransaccionBlockchain"
echo "==========================================================="

# Créer un événement INGRESO
echo "📦 Création d'un événement INGRESO..."
INGRESO_RESPONSE=$(curl -s -X POST "$TRANSACCION_API" \
  -H "Content-Type: application/json" \
  -d '{
    "tipoEvento": "INGRESO",
    "idProducto": "PROD123",
    "lote": "L001",
    "datosEvento": {
      "cantidad": 100,
      "nombreProducto": "Paracetamol 500mg",
      "fabricante": "Laboratorio ABC",
      "proveedor": "PROV001",
      "fechaVencimiento": "2025-12-31",
      "numeroLote": "L001"
    },
    "actorEmisor": "PROVEEDOR_001",
    "ubicacion": "ALMACEN_PRINCIPAL"
  }')

echo "Réponse INGRESO: $INGRESO_RESPONSE"

# Attendre un peu pour que l'événement soit traité
sleep 2

# Créer un événement EGRESO
echo "📤 Création d'un événement EGRESO..."
EGRESO_RESPONSE=$(curl -s -X POST "$TRANSACCION_API" \
  -H "Content-Type: application/json" \
  -d '{
    "tipoEvento": "EGRESO",
    "idProducto": "PROD123",
    "lote": "L001",
    "datosEvento": {
      "cantidad": 50,
      "destino": "HOSPITAL_001",
      "motivoEgreso": "VENTA",
      "numeroFactura": "F001234"
    },
    "actorEmisor": "DISTRIBUIDOR_001",
    "ubicacion": "ALMACEN_PRINCIPAL"
  }')

echo "Réponse EGRESO: $EGRESO_RESPONSE"

# Attendre que les événements soient traités
sleep 3

echo ""
echo "🔍 Étape 2: Test des endpoints HistorialBlockchain avec données réelles"
echo "======================================================================"

# Test 1: Obtenir l'historique du produit
echo "1. Obtenir l'historique du produit PROD123:"
curl -s "$HISTORIAL_API/PROD123" | head -5
echo -e "\n"

# Test 2: Obtenir les événements du produit
echo "2. Obtenir les événements du produit PROD123:"
curl -s "$HISTORIAL_API/PROD123/events?page=1&limit=10" | head -10
echo -e "\n"

# Test 3: Filtrer les événements par type
echo "3. Obtenir seulement les événements INGRESO:"
curl -s "$HISTORIAL_API/PROD123/events?tipo=INGRESO&page=1&limit=5" | head -10
echo -e "\n"

# Test 4: Vérifier un événement spécifique (utiliser l'ID de l'événement créé)
# Extraire l'ID de l'événement de la réponse INGRESO
EVENT_ID=$(echo "$INGRESO_RESPONSE" | grep -o '"idEvento":"[^"]*' | cut -d'"' -f4)
if [ ! -z "$EVENT_ID" ]; then
    echo "4. Vérifier l'événement $EVENT_ID:"
    curl -s "$HISTORIAL_API/PROD123/verify/$EVENT_ID" | head -5
    echo -e "\n"
fi

# Test 5: Reconstruire l'historique
echo "5. Reconstruire l'historique du produit PROD123:"
RECONSTRUCT_RESPONSE=$(curl -s -X POST "$HISTORIAL_API/reconstruir" \
  -H "Content-Type: application/json" \
  -d '{
    "idProducto": "PROD123",
    "lote": "L001",
    "force": true
  }')
echo "$RECONSTRUCT_RESPONSE" | head -5
echo -e "\n"

# Test 6: Obtenir le statut d'une tâche (si une tâche a été créée)
TASK_ID=$(echo "$RECONSTRUCT_RESPONSE" | grep -o '"taskId":"[^"]*' | cut -d'"' -f4)
if [ ! -z "$TASK_ID" ]; then
    echo "6. Obtenir le statut de la tâche $TASK_ID:"
    sleep 2  # Attendre que la tâche soit terminée
    curl -s "$HISTORIAL_API/tasks/$TASK_ID" | head -5
    echo -e "\n"
fi

# Test 7: Lister les inconsistances
echo "7. Lister les inconsistances:"
curl -s "$HISTORIAL_API/inconsistencies?page=1&limit=5" | head-10
echo -e "\n"

echo "✅ Tests d'intégration terminés!"
echo "📊 Vérifiez les logs des services pour plus de détails sur le traitement des événements."
