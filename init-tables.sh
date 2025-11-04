#!/bin/bash

# Script d'initialisation des tables DynamoDB pour les tests
# blockchain-history-medisupply

echo "🔧 Initialisation des tables DynamoDB..."

# Configuration
ENDPOINT="http://localhost:8000"
REGION="us-east-1"

# Fonction pour créer une table
create_table() {
    local table_name=$1
    local key_schema=$2
    local attribute_definitions=$3
    
    echo "📋 Création de la table: $table_name"
    
    aws dynamodb create-table \
        --table-name "$table_name" \
        --key-schema "$key_schema" \
        --attribute-definitions "$attribute_definitions" \
        --billing-mode PAY_PER_REQUEST \
        --endpoint-url "$ENDPOINT" \
        --region "$REGION" \
        --no-cli-pager
    
    echo "✅ Table $table_name créée"
}

# Attendre que DynamoDB soit prêt
echo "⏳ Attente de DynamoDB local..."
while ! curl -s "$ENDPOINT" > /dev/null; do
    sleep 1
done
echo "✅ DynamoDB local est prêt"

# Table historial_transparencia
# Clé primaire: idProducto (String) + lote (String)
create_table "historial_transparencia" \
    'AttributeName=idProducto,KeyType=HASH AttributeName=lote,KeyType=RANGE' \
    'AttributeName=idProducto,AttributeType=S AttributeName=lote,AttributeType=S'

# Table evento_verificado  
# Clé primaire: idProducto (String) + idEvento (String)
create_table "evento_verificado" \
    'AttributeName=idProducto,KeyType=HASH AttributeName=idEvento,KeyType=RANGE' \
    'AttributeName=idProducto,AttributeType=S AttributeName=idEvento,AttributeType=S'

echo ""
echo "🎉 Toutes les tables ont été créées avec succès !"
echo ""
echo "📊 Interface DynamoDB Admin: http://localhost:8001"
echo "🔍 Pour vérifier les tables:"
echo "   aws dynamodb list-tables --endpoint-url http://localhost:8000 --region us-east-1"
echo ""
