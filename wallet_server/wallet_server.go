package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"

	"github.com/hirasawayuki/block_chain/block"
	"github.com/hirasawayuki/block_chain/utils"
	"github.com/hirasawayuki/block_chain/wallet"
)

const tempDir = "wallet_server/templates/"

// WalletServer is wallet server
type WalletServer struct {
	port    uint16
	gateway string
}

// NewWalletServer is returns a WalletServer struct
func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{
		port:    port,
		gateway: gateway,
	}
}

// Port is returns a WalletServer port
func (ws *WalletServer) Port() uint16 {
	return ws.port
}

// Gateway is returns a WalletServer gateway
func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

// Index is handler function that is response index.html
func (ws *WalletServer) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index.html"))
		t.Execute(w, "")
	default:
		log.Println("ERROR: Invalid HTTP Method")
	}
}

// Wallet is handler function that is response wallet.Wallet data.
func (ws *WalletServer) Wallet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

// CreateTransaction is handler function that is create Transaction
func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t wallet.TransactionRequest
		if err := decoder.Decode(&t); err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		if !t.Validate() {
			log.Println("ERROR: missing field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Println("ERROR: parse error")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		value32 := float32(value)
		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &block.TransactionRequest{
			SenderBlockchainAddress:    t.SenderBlockchainAddress,
			RecipientBlockchainAddress: t.RecipientBlockchainAddress,
			SenderPublicKey:            t.SenderPublicKey,
			Value:                      &value32,
			Signature:                  &signatureStr,
		}
		m, _ := json.Marshal(bt)
		buf := bytes.NewBuffer(m)
		resp, err := http.Post(ws.Gateway()+"/transactions", "application/json", buf)
		if err != nil {
			log.Printf("ERROR %v\n", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if resp.StatusCode == 201 {
			log.Printf("Success")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, string(utils.JsonStatus("succ")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("fail")))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method (CreateTransaction)")
	}
}

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		blockchainAddress := r.URL.Query().Get("blockchain_address")
		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())
		client := &http.Client{}
		bcsReq, _ := http.NewRequest("GET", endpoint, nil)
		q := bcsReq.URL.Query()
		q.Add("blockchain_address", blockchainAddress)
		bcsReq.URL.RawQuery = q.Encode()
		bcsResp, err := client.Do(bcsReq)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if bcsResp.StatusCode == 200 {
			decoder := json.NewDecoder(bcsResp.Body)
			var bar block.AmountResponse
			err := decoder.Decode(&bar)
			if err != nil {
				log.Printf("ERROR: %s\n", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}
			m, err := json.Marshal(struct {
				Message string  `json:"message,omitempty"`
				Amount  float32 `json:"amount,omitempty"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			})
			io.WriteString(w, string(m))
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

// Run is start WalletServer
func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
