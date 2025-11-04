#!/bin/bash

# Script d'initialisation des tables DynamoDB pour les tests
# blockchain-history-medisupply

set -e

echo "🔧 Initialisation des tables DynamoDB pour les tests"
echo "===================================================="

# Configuration
DYNAMODB_ENDPOINT="http://localhost:8000"
export AWS_ACCESS_KEY_ID="test-key-id"
export AWS_SECRET_ACCESS_KEY="test-secret-key"
export AWS_DEFAULT_REGION="us-east-1"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Fonction pour attendre que DynamoDB soit prêt
wait_for_dynamodb() {
    echo -e "${YELLOW}⏳ Attente de DynamoDB local...${NC}"
    local attempts=0
    local max_attempts=30
    
    while [ $attempts -lt $max_attempts ]; do
        if curl -s $DYNAMODB_ENDPOINT > /dev/null 2>&1; then
            echo -e "${GREEN}✅ DynamoDB local est prêt${NC}"
            return 0
        fi
        echo "   Tentative $((attempts + 1))/$max_attempts..."
        sleep 2
        attempts=$((attempts + 1))
    done
    
    echo -e "${RED}❌ Timeout: DynamoDB local non disponible${NC}"
    exit 1
}

# Fonction pour créer une table
create_table() {
    local table_name=$1
    local key_schema=$2
    local attribute_definitions=$3
    
    echo -e "${YELLOW}📋 Création de la table ${table_name}...${NC}"
    
    aws dynamodb create-table \
        --table-name "$table_name" \
        --attribute-definitions $attribute_definitions \
        --key-schema $key_schema \
        --billing-mode PAY_PER_REQUEST \
        --endpoint-url $DYNAMODB_ENDPOINT \
        --no-cli-pager > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Table ${table_name} créée avec succès${NC}"
    else
        echo -e "${RED}⚠️  Table ${table_name} existe déjà ou erreur de création${NC}"
    fi
}

# Fonction pour insérer des données de test
insert_test_data() {
    echo -e "${YELLOW}📝 Insertion de données de test...${NC}"
    
    # Données de test pour historial_transparencia
    aws dynamodb put-item \
        --table-name historial_transparencia \
        --item '{
            "idProducto": {"S": "PROD123"},
            "lote": {"S": "L001"},
            "nombreProducto": {"S": "Paracetamol 500mg"},
            "fabricante": {"S": "PharmaCorp"},
            "estadoActual": {"S": "Conforme"},
            "validacionBlockchain": {"BOOL": true},
            "ultimoCheck": {"S": "2024-11-04T10:00:00Z"},
            "metadata": {"M": {
                "categoria": {"S": "analgesico"},
                "origen": {"S": "nacional"}
            }},
            "createdAt": {"S": "2024-10-01T08:00:00Z"},
            "updatedAt": {"S": "2024-11-04T10:00:00Z"}
        }' \
        --endpoint-url $DYNAMODB_ENDPOINT \
        --no-cli-pager > /dev/null
    
    # Données de test pour evento_verificado
    aws dynamodb put-item \
        --table-name evento_verificado \
        --item '{
            "idProducto": {"S": "PROD123"},
            "idEvento": {"S": "EVT456"},
            "tipoEvento": {"S": "INGRESO"},
            "fecha": {"S": "2024-10-01T08:00:00Z"},
            "ubicacion": {"S": "Almacén Central"},
            "datosEvento": {"M": {
                "cantidad": {"N": "100"},
                "lote": {"S": "L001"},
                "proveedor": {"S": "PROV001"}
            }},
            "hashEvento": {"S": "0xabc123def456..."},
            "referenciaBlockchain": {"S": "0x123abc456def..."},
            "resultadoVerificacion": {"S": "OK"},
            "observaciones": {"S": "Evento verificado correctement"},
            "createdAt": {"S": "2024-10-01T08:00:00Z"}
        }' \
        --endpoint-url $DYNAMODB_ENDPOINT \
        --no-cli-pager > /dev/null
    
    # Deuxième événement de test
    aws dynamodb put-item \
        --table-name evento_verificado \
        --item '{
            "idProducto": {"S": "PROD123"},
            "idEvento": {"S": "EVT789"},
            "tipoEvento": {"S": "EGRESO"},
            "fecha": {"S": "2024-10-15T14:20:00Z"},
            "ubicacion": {"S": "Pharmacie Centrale"},
            "datosEvento": {"M": {
                "cantidad": {"N": "50"},
                "destino": {"S": "HOSPITAL_001"}
            }},
            "hashEvento": {"S": "0x456def789abc..."},
            "referenciaBlockchain": {"S": "0x456def789abc..."},
            "resultadoVerificacion": {"S": "OK"},
            "observaciones": {"S": "Distribution vérifiée"},
            "createdAt": {"S": "2024-10-15T14:20:00Z"}
        }' \
        --endpoint-url $DYNAMODB_ENDPOINT \
        --no-cli-pager > /dev/null
    
    echo -e "${GREEN}✅ Données de test insérées${NC}"
}

# Fonction principale
main() {
    # Vérifier si AWS CLI est installé
    if ! command -v aws &> /dev/null; then
        echo -e "${RED}❌ AWS CLI non installé. Installez-le d'abord:${NC}"
        echo "   sudo apt install awscli  # Ubuntu/Debian"
        echo "   brew install awscli      # macOS"
        exit 1
    fi
    
    # Attendre DynamoDB
    wait_for_dynamodb
    
    # Créer les tables
    echo -e "\n${YELLOW}📊 Création des tables DynamoDB...${NC}"
    
    create_table "historial_transparencia" \
        "AttributeName=idProducto,KeyType=HASH AttributeName=lote,KeyType=RANGE" \
        "AttributeName=idProducto,AttributeType=S AttributeName=lote,AttributeType=S"
    
    create_table "evento_verificado" \
        "AttributeName=idProducto,KeyType=HASH AttributeName=idEvento,KeyType=RANGE" \
        "AttributeName=idProducto,AttributeType=S AttributeName=idEvento,AttributeType=S"
    
    # Attendre que les tables soient actives
    echo -e "${YELLOW}⏳ Attente de l'activation des tables...${NC}"
    sleep 3
    
    # Insérer des données de test
    insert_test_data
    
    # Vérifier les tables créées
    echo -e "\n${YELLOW}🔍 Vérification des tables...${NC}"
    TABLES=$(aws dynamodb list-tables --endpoint-url $DYNAMODB_ENDPOINT --output text --query 'TableNames' 2>/dev/null)
    
    if echo "$TABLES" | grep -q "historial_transparencia" && echo "$TABLES" | grep -q "evento_verificado"; then
        echo -e "${GREEN}✅ Toutes les tables ont été créées avec succès${NC}"
    else
        echo -e "${RED}❌ Erreur dans la création des tables${NC}"
        exit 1
    fi
    
    echo -e "\n${GREEN}🎉 Initialisation terminée avec succès!${NC}"
    echo -e "\n${YELLOW}📱 Accès aux interfaces:${NC}"
    echo "   - DynamoDB Admin: http://localhost:8001"
    echo "   - Kafka UI: http://localhost:8090"
    echo ""
}

# Exécuter le script
main "$@"
