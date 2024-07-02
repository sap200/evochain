package blockchainserver

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/sap200/evochain/blockchain"
	"github.com/sap200/evochain/constants"
)

type BlockchainServer struct {
	Port          uint64                       `json:"port"`
	BlockchainPtr *blockchain.BlockchainStruct `json:"blockchain"`
}

func NewBlockchainServer(port uint64, blockchainPtr *blockchain.BlockchainStruct) *BlockchainServer {
	bcs := new(BlockchainServer)
	bcs.Port = port
	bcs.BlockchainPtr = blockchainPtr

	return bcs
}

func (bcs *BlockchainServer) GetBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		io.WriteString(w, bcs.BlockchainPtr.ToJson())
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetBalance(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		addr := req.URL.Query().Get("address")
		x := struct {
			Balance uint64 `json:"balance"`
		}{
			bcs.BlockchainPtr.CalculateTotalCrypto(addr),
		}

		mBalance, err := json.Marshal(x)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, string(mBalance))
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetAllNonRewardedTxns(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		txnList := bcs.BlockchainPtr.GetAllTxns()
		byteSlice, err := json.Marshal(txnList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, string(byteSlice))
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) SendTxnToTheBlockchain(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		request, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		var newTxn blockchain.Transaction

		err = json.Unmarshal(request, &newTxn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		go bcs.BlockchainPtr.AddTransactionToTransactionPool(&newTxn)

		io.WriteString(w, newTxn.ToJson())
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}

}

func CheckStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		io.WriteString(w, constants.BLOCKCHAIN_STATUS)
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) SendPeersList(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodPost {
		peersMap, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("Error reading Peers")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var peersList map[string]bool
		err = json.Unmarshal(peersMap, &peersList)
		if err != nil {
			log.Println("Error Unmarshalling the Peers")

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		go bcs.BlockchainPtr.UpdatePeers(peersList)
		res := map[string]string{}
		res["status"] = "success"
		x, err := json.Marshal(res)
		if err != nil {
			log.Println("Error while marshalling the http Response to be sent.")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, string(x))
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) FetchLastNBlocks(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		blocks := bcs.BlockchainPtr.Blocks
		blockchain1 := new(blockchain.BlockchainStruct)
		if len(blocks) < constants.FETCH_LAST_N_BLOCKS {
			blockchain1.Blocks = blocks
		} else {
			blockchain1.Blocks = blocks[len(blocks)-constants.FETCH_LAST_N_BLOCKS:]
		}

		io.WriteString(w, blockchain1.ToJson())
	} else {
		http.Error(w, "Invalid Method", http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Start() {
	http.HandleFunc("/", bcs.GetBlockchain)
	http.HandleFunc("/balance", bcs.GetBalance)
	http.HandleFunc("/get_all_non_rewarded_txns", bcs.GetAllNonRewardedTxns)
	http.HandleFunc("/send_txn", bcs.SendTxnToTheBlockchain)
	http.HandleFunc("/send_peers_list", bcs.SendPeersList)
	http.HandleFunc("/check_status", CheckStatus)
	http.HandleFunc("/fetch_last_n_blocks", bcs.FetchLastNBlocks)
	log.Println("Launching webserver at port :", bcs.Port)
	err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(bcs.Port)), nil)
	if err != nil {
		panic(err)
	}
}
