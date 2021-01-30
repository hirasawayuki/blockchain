package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hirasawayuki/block_chain/utils"
	"golang.org/x/crypto/ripemd160"
)

// Wallet is struct that dedicated to cryptocurrencies
type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

// NewWallet is return Wallet struct with public key and private key
func NewWallet() *Wallet {
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 0 - Having a private ECDSA key
	// 18e14a7b6a307f426a94f8114701e7c8e774e7f9a47e2c2035db29a206321725
	w.privateKey = privateKey

	// 1 - Take the corresponding public key generated with it (33 bytes, 1 byte 0x02 (y-coord is even), and 32 bytes corresponding to X coordinate)
	// 0250863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352
	w.publicKey = &w.privateKey.PublicKey

	// 2 - Perform SHA-256 hashing on the public key
	// 0b7c28c9b7290c98d7438e70b3d3f7c848fbd7d1dc194ff83f4f7cc9b1378e98
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	// 3 - Perform RIPEMD-160 hashing on the result of SHA-256
	// f54a5851e9372b87810a8e60cdd2e7cfd80b6e31
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	// 4 - Add version byte in front of RIPEMD-160 hash (0x00 for Main Network)
	// 00f54a5851e9372b87810a8e60cdd2e7cfd80b6e31
	// (note that below steps are the Base58Check encoding, which has multiple library options available implementing it)
	vb4 := make([]byte, 21)
	copy(vb4[1:], digest3)

	// 5 - Perform SHA-256 hash on the extended RIPEMD-160 result
	// ad3c854da227c7e99c4abfad4ea41d71311160df2e415e713318c70d67c6b41c

	h5 := sha256.New()
	h5.Write(vb4)
	digest5 := h5.Sum(nil)

	// 6 - Perform SHA-256 hash on the result of the previous SHA-256 hash
	// c7f18fe8fcbed6396741e58ad259b5cb16b7fd7f041904147ba1dcffabf747fd
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)

	// 7 - Take the first 4 bytes of the second SHA-256 hash. This is the address checksum
	// c7f18fe8
	checksum := digest6[:4]

	// 8 - Add the 4 checksum bytes from stage 7 at the end of extended RIPEMD-160 hash from stage 4. This is the 25-byte binary Bitcoin Address.
	// 00f54a5851e9372b87810a8e60cdd2e7cfd80b6e31c7f18fe8
	h8 := make([]byte, 25)
	copy(h8[:21], vb4)
	copy(h8[21:], checksum)

	// 9 - Convert the result from a byte string into a base58 string using Base58Check encoding. This is the most commonly used Bitcoin Address format
	// 1PMycacnJaSqwwJqjawXBErnLsZ7RkXUAs
	address := base58.Encode(h8)
	w.blockchainAddress = address
	return w
}

// PrivateKey is returns a Wallet private key
func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

// PrivateKeyStr is returns string that convert private key data from []byte to string
func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

// PublicKey is returns a Wallet publick key
func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

// PublicKeyStr is returns string that convert publick key data(X, Y) from []byte to string
func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

// BlockchainAddress is returns a blockchain address strings
func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

// MarshalJSON returns the JSON encoding Wallet struct.
func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey        string `json:"private_key,omitempty"`
		PublicKey         string `json:"public_key,omitempty"`
		BlockchainAddress string `json:"blockchain_address,omitempty"`
	}{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.blockchainAddress,
	})
}

// Transaction is struct of transaction with senderPrivateKey, senderPublickKey, senderBlockchainAddress, recipientBlockchainAddress, value
type Transaction struct {
	senderPrivateKey           *ecdsa.PrivateKey
	senderPublickKey           *ecdsa.PublicKey
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

// MarshalJSON is returns a struct with sender_blockchain_address, recipient_blockchain_address, value
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderBlockchainAddress    string  `json:"sender_blockchain_address,omitempty"`
		RecipientBlockchainAddress string  `json:"recipient_blockchain_address,omitempty"`
		Value                      float32 `json:"value,omitempty"`
	}{
		SenderBlockchainAddress:    t.senderBlockchainAddress,
		RecipientBlockchainAddress: t.recipientBlockchainAddress,
		Value:                      t.value,
	})
}

// NewTransaction is returns a pointer that Transaction struct
func NewTransaction(privateKey *ecdsa.PrivateKey, publickKey *ecdsa.PublicKey, sender string, recipient string, value float32) *Transaction {
	return &Transaction{privateKey, publickKey, sender, recipient, value}
}

// GenerateSignature is returns a Signature struct
func (t *Transaction) GenerateSignature() *utils.Signature {
	m, _ := json.Marshal(t)
	h := sha256.Sum256(m)
	r, s, _ := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	return &utils.Signature{R: r, S: s}
}

type TransactionRequest struct {
	SenderPrivateKey           *string `json:"sender_private_key,omitempty"`
	SenderBlockchainAddress    *string `json:"sender_blockchain_address,omitempty"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address,omitempty"`
	SenderPublicKey            *string `json:"sender_public_key,omitempty"`
	Value                      *string `json:"value,omitempty"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderPrivateKey == nil ||
		tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil {
		return false
	}
	return true
}
