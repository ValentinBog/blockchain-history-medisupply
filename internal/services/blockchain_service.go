package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"

	"github.com/edinfamous/historial-blockchain/internal/models"
)

// BlockchainService gère les interactions avec la blockchain
type BlockchainService struct {
	client    *ethclient.Client
	rpcURL    string
	timeout   time.Duration
	maxRetries int
}

// NewBlockchainService crée une nouvelle instance de BlockchainService
func NewBlockchainService(rpcURL string, timeout time.Duration, maxRetries int) (*BlockchainService, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la blockchain: %w", err)
	}

	return &BlockchainService{
		client:     client,
		rpcURL:     rpcURL,
		timeout:    timeout,
		maxRetries: maxRetries,
	}, nil
}

// VerificarIntegridad vérifie l'intégrité d'un événement contre la blockchain
func (bs *BlockchainService) VerificarIntegridad(ctx context.Context, evento *models.EventoVerificado) error {
	if evento.ReferenciaBlockchain == "" {
		return fmt.Errorf("référence blockchain manquante")
	}

	// Créer contexte avec timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, bs.timeout)
	defer cancel()

	// Récupérer la transaction de la blockchain
	txHash := common.HexToHash(evento.ReferenciaBlockchain)
	
	var err error
	
	// Retry avec backoff exponentiel
	for i := 0; i < bs.maxRetries; i++ {
		_, err = bs.client.TransactionReceipt(ctxWithTimeout, txHash)
		if err == nil {
			break
		}
		
		if i < bs.maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	
	if err != nil {
		evento.ResultadoVerificacion = models.VerificacionNotFound
		evento.Observaciones = fmt.Sprintf("Transaction not found: %v", err)
		return fmt.Errorf("transaction non trouvée: %w", err)
	}

	// Calculer le hash local
	hashLocal, err := bs.calcularHashLocal(evento.DatosEvento)
	if err != nil {
		evento.ResultadoVerificacion = models.VerificacionHashMismatch
		evento.Observaciones = fmt.Sprintf("Erreur calcul hash local: %v", err)
		return fmt.Errorf("erreur calcul hash local: %w", err)
	}

	// Comparer les hashs
	if hashLocal != evento.HashEvento {
		evento.ResultadoVerificacion = models.VerificacionHashMismatch
		evento.Observaciones = fmt.Sprintf("Hash mismatch: local=%s, événement=%s", hashLocal, evento.HashEvento)
		return fmt.Errorf("hash mismatch")
	}

	evento.ResultadoVerificacion = models.VerificacionOK
	evento.Observaciones = "Vérification réussie"
	return nil
}

// calcularHashLocal calcule le hash local des données d'événement
func (bs *BlockchainService) calcularHashLocal(datosEvento map[string]interface{}) (string, error) {
	// Convertir en JSON pour calculer le hash
	jsonData, err := json.Marshal(datosEvento)
	if err != nil {
		return "", fmt.Errorf("erreur marshalling JSON: %w", err)
	}

	// Calculer SHA256
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// GetTransactionByHash récupère une transaction par son hash
func (bs *BlockchainService) GetTransactionByHash(ctx context.Context, txHash string) (interface{}, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, bs.timeout)
	defer cancel()

	hash := common.HexToHash(txHash)
	
	for i := 0; i < bs.maxRetries; i++ {
		tx, err := bs.client.TransactionReceipt(ctxWithTimeout, hash)
		if err == nil {
			return tx, nil
		}
		
		if i < bs.maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	
	return nil, fmt.Errorf("impossible de récupérer la transaction après %d tentatives", bs.maxRetries)
}

// VerificarConexion vérifie la connexion à la blockchain
func (bs *BlockchainService) VerificarConexion(ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := bs.client.NetworkID(ctxWithTimeout)
	if err != nil {
		return fmt.Errorf("impossible de se connecter à la blockchain: %w", err)
	}

	log.Println("✅ Connexion blockchain vérifiée")
	return nil
}

// Close ferme la connexion blockchain
func (bs *BlockchainService) Close() {
	if bs.client != nil {
		bs.client.Close()
	}
}
