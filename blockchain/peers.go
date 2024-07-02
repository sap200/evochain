package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sap200/evochain/constants"
)

func SyncBlockchain(address string) (*BlockchainStruct, error) {
	log.Println("Started syncing blockchain from node:", address)
	ourURL := fmt.Sprintf("%s/", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		return nil, err
	}

	// read the body of the response here
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bs BlockchainStruct
	err = json.Unmarshal(data, &bs)
	if err != nil {
		return nil, err
	}

	log.Println("Finished syncing blockchain from node:", address)

	return &bs, nil
}

func (bc *BlockchainStruct) UpdatePeers(peersList map[string]bool) {
	mutex.Lock()
	defer mutex.Unlock()

	log.Println("Updating Peers List..", peersList)
	bc.Peers = peersList

	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) SendPeersList(address string) {
	data := bc.PeersToJson()
	ourURL := fmt.Sprintf("%s/send_peers_list", address)
	http.Post(ourURL, "application/json", bytes.NewBuffer(data))
}

func (bc *BlockchainStruct) CheckStatus(address string) bool {
	ourURL := fmt.Sprintf("%s/check_status", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		log.Println(err)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	return string(data) == constants.BLOCKCHAIN_STATUS
}

func (bc *BlockchainStruct) BroadcastPeerList() {
	for peer, status := range bc.Peers {
		if peer != bc.Address && status {
			bc.SendPeersList(peer)
			time.Sleep(constants.PEER_BROADCAST_PAUSE_TIME * time.Second)
		}
	}
}

func (bc *BlockchainStruct) DialAndUpdatePeers() {
	for {
		log.Println("Pinging Peers", bc.Peers)
		newList := bc.Peers

		for peer := range newList {
			if peer != bc.Address {
				newList[peer] = bc.CheckStatus(peer)
			} else {
				newList[peer] = true
			}
		}

		// update our peers List
		bc.UpdatePeers(newList)
		log.Println("Updated Peer status : ", bc.Peers)

		// broadcast our new peers list
		bc.BroadcastPeerList()

		time.Sleep(constants.PEER_PING_PAUSE_TIME * time.Second)
	}
}

// For transaction

func (bc *BlockchainStruct) SendTxnToThePeer(address string, txn *Transaction) {
	data := txn.ToJson()
	ourURL := fmt.Sprintf("%s/send_txn", address)
	http.Post(ourURL, "application/json", strings.NewReader(data))
}

func (bc *BlockchainStruct) BroadcastTransaction(txn *Transaction) {
	for peer, status := range bc.Peers {
		if peer != bc.Address && status {
			log.Println("Broadcasting transaction to the peer:", peer, "Transaction:", txn.ToJson())
			bc.SendTxnToThePeer(peer, txn)
			time.Sleep(constants.TXN_BROADCAST_PAUSE_TIME * time.Second)
		}
	}
}

func FetchLastNBlocks(address string) (*BlockchainStruct, error) {
	log.Println("Fetching last", constants.FETCH_LAST_N_BLOCKS, "blocks")
	ourURL := fmt.Sprintf("%s/fetch_last_n_blocks", address)
	resp, err := http.Get(ourURL)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nbc BlockchainStruct
	err = json.Unmarshal(data, &nbc)
	if err != nil {
		return nil, err
	}

	return &nbc, nil
}

func verifyLastNBlocks(chain []*Block) bool {
	if chain[0].BlockNumber != 0 && chain[0].Hash()[2:2+constants.MINING_DIFFICULTY] != strings.Repeat("0", constants.MINING_DIFFICULTY) {
		log.Println("Chain verification failed for block", chain[0].BlockNumber, "hash", chain[0].Hash())
		return false
	}

	for i := 1; i < len(chain); i++ {
		if chain[i-1].Hash() != chain[i].PrevHash {
			log.Println("Failed to verify prevHash for block number", chain[i].BlockNumber)
			return false
		}

		if chain[i].Hash()[2:2+constants.MINING_DIFFICULTY] != strings.Repeat("0", constants.MINING_DIFFICULTY) {
			log.Println("Chain verification failed for block", chain[0].BlockNumber, "hash", chain[0].Hash())
			return false
		}
	}

	return true
}

func (bc *BlockchainStruct) UpdateBlockchain(chain []*Block) {
	mutex.Lock()
	defer mutex.Unlock()

	blocks := []*Block{}
	initIdx := chain[0].BlockNumber
	log.Println("Updating our blockchain from block number", initIdx)
	blocks = append(blocks, bc.Blocks[:initIdx]...)
	blocks = append(blocks, chain...)

	bc.Blocks = blocks

	// update the transaction pool
	found := map[string]bool{}
	for _, txn := range bc.TransactionPool {
		found[txn.TransactionHash] = false
	}

	for _, block := range chain {
		for _, txn := range block.Transactions {
			_, ok := found[txn.TransactionHash]
			if ok {
				found[txn.TransactionHash] = true
			}
		}
	}

	newTxnPool := []*Transaction{}
	for _, txn := range bc.TransactionPool {
		if !found[txn.TransactionHash] {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	bc.TransactionPool = newTxnPool

	// save the blockchain in the database
	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) RunConsensus() {

	for {
		log.Println("Starting the consensus algorithm...")
		longestChain := bc.Blocks
		lengthOfTheLongestChain := bc.Blocks[len(bc.Blocks)-1].BlockNumber + 1
		longestChainIsOur := true
		for peer, status := range bc.Peers {
			if peer != bc.Address && status {
				bc1, err := FetchLastNBlocks(peer)
				if err != nil {
					log.Println("Error while  fetching last n blocks from peer:", peer, "Error:", err.Error())
					continue
				}

				lengthOfTheFetchedChain := bc1.Blocks[len(bc1.Blocks)-1].BlockNumber + 1
				if lengthOfTheFetchedChain > lengthOfTheLongestChain {
					longestChain = bc1.Blocks
					lengthOfTheLongestChain = lengthOfTheFetchedChain
					longestChainIsOur = false
				}
			}
		}

		if longestChainIsOur {
			log.Println("My chain is longest, thus I am not updating my blockchain")
			time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
			continue
		}

		if verifyLastNBlocks(longestChain) {
			// stop the Mining until updation
			bc.MiningLocked = true
			bc.UpdateBlockchain(longestChain)
			// restart the Mining as updation is complete
			bc.MiningLocked = false
			log.Println("Updation of Blockchain complete !!!")
		} else {
			log.Println("Chain Verification Failed, Hence not updating my blockchain")
		}

		time.Sleep(constants.CONSENSUS_PAUSE_TIME * time.Second)
	}

}
