package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"../utils"
)

const (
	// MiningDifficulty is the difficuluty of mining
	MiningDifficulty = 3
	// MiningSender is Blockchain network address
	MiningSender = "THE BLOCKCHAIN"
	// MiningReward is a mining reward
	MiningReward = 1.0
)

// Block is a structure with nonce, previousHash, timestamp, transactions
type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

// Hash convert Block to SHA256 []byte and returns []byte
func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

// MarshalJSON is returns a struct
func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `j:"timestamp"`
		Nonce        int            `j:"nonce"`
		PreviousHash [32]byte       `j:"previous_hash"`
		Transactions []*Transaction `j:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: b.previousHash,
		Transactions: b.transactions,
	})
}

// NewBlock returns a Block structure pointer.
func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	b := new(Block)
	b.nonce = nonce
	b.previousHash = previousHash
	b.timestamp = time.Now().UnixNano()
	b.transactions = transactions
	return b
}

// Print is formats and outputs
func (b *Block) Print() {
	fmt.Printf("timestamp:     %d\n", b.timestamp)
	fmt.Printf("nonce:         %d\n", b.nonce)
	fmt.Printf("previousHash:  %x\n", b.previousHash)

	for _, t := range b.transactions {
		t.Print()
	}
}

// Blockchain is a struct with transactionsPool, chain
type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
}

// NewBlockChain returns a Blockchain struct
func NewBlockChain(blockchainAddress string, port uint16) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

// CreateBlock is create Block and append chain.
// returns a Block
func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.transactionPool = []*Transaction{}
	bc.chain = append(bc.chain, b)
	return b
}

// LastBlock returns last Block in Blockchain
func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

// Print is output chain.
func (bc *Blockchain) Print() {
	for i, b := range bc.chain {
		fmt.Println("#############################")
		fmt.Printf("chain:         %d\n", i)
		b.Print()
	}
}

// AddTransaction is create Transaction and add BlockChain struct
func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublickKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)
	if sender == MiningSender {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}
	if bc.VerifyTransactionSignature(senderPublickKey, s, t) {
		// if bc.CaluculateTotalAmount(sender) < value {
		// 	log.Println("ERROR: Not enough balance in a wallet")
		// 	return false
		// }
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}
	log.Println("ERROR: Verify Transaction")
	return false
}

// VerifyTransactionSignature is verify transaction
func (bc *Blockchain) VerifyTransactionSignature(senderPublickKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublickKey, h[:], s.R, s.S)
}

// CopyTransactionPool is copy TransactionPool and returns a Transaction pointer slice.
func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)

	for _, t := range bc.transactionPool {
		transactions = append(transactions, NewTransaction(t.senderBlockchainAddress, t.recipientBlockchainAddress, t.value))
	}
	return transactions
}

// ValidProof is checks that the first difficuluty(3) digits of the hash value are 0
func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficuluty int) bool {
	zeros := strings.Repeat("0", difficuluty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficuluty] == zeros
}

// ProofOfWork is find a nonce where ValidProof is true
func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MiningDifficulty) {
		nonce++
	}
	return nonce
}

// Mining is add transactions and pay miner for mining.
func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(MiningSender, bc.blockchainAddress, MiningReward, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	fmt.Println("action=mining, status=success")
	return true
}

// CaluculateTotalAmount is caluculate the wallet balance that matches the blockchain address
func (bc *Blockchain) CaluculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			if t.senderBlockchainAddress == blockchainAddress {
				totalAmount -= t.value
			}
			if t.recipientBlockchainAddress == blockchainAddress {
				totalAmount += t.value
			}
		}
	}
	return totalAmount
}

// Transaction is struct with senderBlockchainAddress, recipientBlockchainAddress, value
type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

// NewTransaction is return a Transaction struct pointer
func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

// Print is format Transaction struct and output
func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf("senderBlockchainAddress:     %s\n", t.senderBlockchainAddress)
	fmt.Printf("recipientBlockchainAddress:  %s\n", t.recipientBlockchainAddress)
	fmt.Printf("value:                       %.1f\n", t.value)
}

// MarshalJSON is marshal Transaction
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address,omitempty"`
		Recipient string  `json:"recipient_blockchain_address,omitempty"`
		Value     float32 `json:"value,omitempty"`
	}{
		t.senderBlockchainAddress,
		t.recipientBlockchainAddress,
		t.value,
	})
}
