package walletserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sap200/evochain/blockchain"
	"github.com/sap200/evochain/constants"
	"github.com/sap200/evochain/wallet"
)

type WalletServer struct {
	Port                  uint64 `json:"port"`
	BlockchainNodeAddress string `json:"blockchain_node_addres"`
}

func NewWalletServer(port uint64, blockchainNodeAddress string) *WalletServer {
	ws := new(WalletServer)
	ws.Port = port
	ws.BlockchainNodeAddress = blockchainNodeAddress
	return ws
}

func (ws *WalletServer) CreateNewWallet(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		wallet1, err := wallet.NewWallet()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		x := struct {
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
			Address    string `json:"address"`
		}{
			wallet1.GetPrivateKeyHex(),
			wallet1.GetPublicKeyHex(),
			wallet1.GetAddress(),
		}

		wbs, err := json.Marshal(x)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		io.WriteString(w, string(wbs))

	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func (ws *WalletServer) GetTotalCryptoFromWallet(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		params := url.Values{}
		params.Add("address", req.URL.Query().Get("address"))
		ourURL := fmt.Sprintf("%s?%s", ws.BlockchainNodeAddress+"/balance", params.Encode())
		resp, err := http.Get(ourURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, string(data))

	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (ws *WalletServer) SendTxnToTheBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		privateKey := req.URL.Query().Get("privateKey")

		dataBs, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
		var txn1 blockchain.Transaction
		err = json.Unmarshal(dataBs, &txn1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		wallet1 := wallet.NewWalletFromPrivateKeyHex(privateKey)

		myTxn := blockchain.NewTransaction(wallet1.GetAddress(), txn1.To, txn1.Value, []byte{})
		myTxn.Status = constants.PENDING
		newTxn, err := wallet1.GetSignedTxn(*myTxn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newTxnBs, err := json.Marshal(newTxn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// send it to the blockchain
		resp, err := http.Post(ws.BlockchainNodeAddress+"/send_txn", "application/json", bytes.NewBuffer(newTxnBs))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resultBs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		io.WriteString(w, string(resultBs))
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func (ws *WalletServer) Start() {
	http.HandleFunc("/wallet_balance", ws.GetTotalCryptoFromWallet)
	http.HandleFunc("/create_new_wallet", ws.CreateNewWallet)
	http.HandleFunc("/send_signed_txn", ws.SendTxnToTheBlockchain)
	log.Println("Starting wallet server at port:", ws.Port)
	err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(ws.Port)), nil)
	if err != nil {
		panic(err)
	}
}
