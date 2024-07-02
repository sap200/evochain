package blockchain

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/sap200/evochain/constants"
)

type BlockchainStruct struct {
	TransactionPool []*Transaction  `json:"transaction_pool"`
	Blocks          []*Block        `json:"block_chain"`
	Address         string          `json:"address"`
	Peers           map[string]bool `json:"peers"`
	MiningLocked    bool            `json:"mining_locked"`
}

var mutex sync.Mutex

func NewBlockchain(genesisBlock Block, address string) *BlockchainStruct {
	exists, _ := KeyExists()

	if exists {
		blockchainStruct, err := GetBlockchain()
		if err != nil {
			panic(err.Error())
		}
		return blockchainStruct
	} else {
		blockchainStruct := new(BlockchainStruct)
		blockchainStruct.TransactionPool = []*Transaction{}
		blockchainStruct.Blocks = []*Block{}
		blockchainStruct.Blocks = append(blockchainStruct.Blocks, &genesisBlock)
		blockchainStruct.Address = address
		blockchainStruct.Peers = map[string]bool{}
		blockchainStruct.MiningLocked = false
		err := PutIntoDb(*blockchainStruct)
		if err != nil {
			panic(err.Error())
		}
		return blockchainStruct
	}
}

func NewBlockchainFromSync(bc1 *BlockchainStruct, address string) *BlockchainStruct {
	bc2 := bc1
	bc2.Address = address

	err := PutIntoDb(*bc2)
	if err != nil {
		panic(err.Error())
	}

	return bc2
}

func (bc BlockchainStruct) PeersToJson() []byte {
	nb, _ := json.Marshal(bc.Peers)

	return nb
}

func (bc BlockchainStruct) ToJson() string {
	nb, err := json.Marshal(bc)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (bc *BlockchainStruct) AddBlock(b *Block) {
	mutex.Lock()
	defer mutex.Unlock()

	m := map[string]bool{}
	for _, txn := range b.Transactions {
		m[txn.TransactionHash] = true
	}

	// remove txn from txn pool
	newTxnPool := []*Transaction{}
	for _, txn := range bc.TransactionPool {
		_, ok := m[txn.TransactionHash]
		if !ok {
			newTxnPool = append(newTxnPool, txn)
		}
	}

	bc.TransactionPool = newTxnPool
	bc.Blocks = append(bc.Blocks, b)

	// save the blockchain to our database
	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) appendTransactionToTheTransactionPool(transaction *Transaction) {
	mutex.Lock()
	defer mutex.Unlock()

	bc.TransactionPool = append(bc.TransactionPool, transaction)

	// save the blockchain to our database
	err := PutIntoDb(*bc)
	if err != nil {
		panic(err.Error())
	}
}

func (bc *BlockchainStruct) AddTransactionToTransactionPool(transaction *Transaction) {

	for _, txn := range bc.TransactionPool {
		if txn.TransactionHash == transaction.TransactionHash {
			return
		}
	}

	log.Println("Adding txn to the Transaction pool")

	newTxn := new(Transaction)
	newTxn.From = transaction.From
	newTxn.To = transaction.To
	newTxn.Value = transaction.Value
	newTxn.Data = transaction.Data
	newTxn.Status = transaction.Status
	newTxn.Timestamp = transaction.Timestamp
	newTxn.TransactionHash = transaction.TransactionHash
	newTxn.PublicKey = transaction.PublicKey
	newTxn.Signature = transaction.Signature

	valid1 := transaction.VerifyTxn()

	valid2 := bc.simulatedBalanceCheck(valid1, transaction)

	if valid1 && valid2 {
		transaction.Status = constants.TXN_VERIFICATION_SUCCESS
	} else {
		transaction.Status = constants.TXN_VERIFICATION_FAILURE
	}

	transaction.PublicKey = ""

	bc.appendTransactionToTheTransactionPool(transaction)

	bc.BroadcastTransaction(newTxn)
}

func (bc *BlockchainStruct) simulatedBalanceCheck(valid1 bool, transaction *Transaction) bool {
	balance := bc.CalculateTotalCrypto(transaction.From)
	for _, txn := range bc.TransactionPool {
		if transaction.From == txn.From && valid1 {
			if balance >= txn.Value {
				balance -= txn.Value
			} else {
				break
			}
		}
	}

	return balance >= transaction.Value
}

func (bc *BlockchainStruct) ProofOfWorkMining(minersAddress string) {
	log.Println("Starting to Mine...")
	// calculate the prevHash
	nonce := 0
	for {
		if bc.MiningLocked {
			continue
		}

		prevHash := bc.Blocks[len(bc.Blocks)-1].Hash()

		if bc.MiningLocked {
			continue
		}

		// start with a nonce
		// create a new block
		guessBlock := NewBlock(prevHash, nonce, uint64(len(bc.Blocks)))

		if bc.MiningLocked {
			continue
		}
		// copy the transaction pool
		for _, txn := range bc.TransactionPool {

			if bc.MiningLocked {
				continue
			}

			newTxn := new(Transaction)
			newTxn.Data = txn.Data
			newTxn.From = txn.From
			newTxn.To = txn.To
			newTxn.Status = txn.Status
			newTxn.Timestamp = txn.Timestamp
			newTxn.Value = txn.Value
			newTxn.TransactionHash = txn.TransactionHash
			newTxn.PublicKey = txn.PublicKey
			newTxn.Signature = txn.Signature

			guessBlock.AddTransactionToTheBlock(newTxn)
		}

		if bc.MiningLocked {
			continue
		}

		rewardTxn := NewTransaction(constants.BLOCKCHAIN_ADDRESS, minersAddress, constants.MINING_REWARD, []byte{})
		rewardTxn.Status = constants.SUCCESS
		guessBlock.Transactions = append(guessBlock.Transactions, rewardTxn)

		if bc.MiningLocked {
			continue
		}

		// guess the Hash
		guessHash := guessBlock.Hash()
		desiredHash := strings.Repeat("0", constants.MINING_DIFFICULTY)
		ourSolutionHash := guessHash[2 : 2+constants.MINING_DIFFICULTY]

		if bc.MiningLocked {
			continue
		}

		if ourSolutionHash == desiredHash {

			if !bc.MiningLocked {
				bc.AddBlock(guessBlock)
				log.Println("Mined block number:", guessBlock.BlockNumber)
			}
			nonce = 0
			continue
		}

		nonce++
	}

}

func (bc *BlockchainStruct) CalculateTotalCrypto(address string) uint64 {
	sum := uint64(0)

	for _, blocks := range bc.Blocks {
		for _, txns := range blocks.Transactions {
			if txns.Status == constants.SUCCESS {
				if txns.To == address {
					sum += txns.Value
				} else if txns.From == address {
					sum -= txns.Value
				}
			}
		}
	}
	return sum
}

func (bc *BlockchainStruct) GetAllTxns() []Transaction {

	nTxns := []Transaction{}

	for i := len(bc.TransactionPool) - 1; i >= 0; i-- {
		nTxns = append(nTxns, *bc.TransactionPool[i])
	}

	txns := []Transaction{}

	for _, blocks := range bc.Blocks {
		for _, txn := range blocks.Transactions {
			if txn.From != constants.BLOCKCHAIN_ADDRESS {
				txns = append(txns, *txn)
			}
		}
	}
	for i := len(txns) - 1; i >= 0; i-- {
		nTxns = append(nTxns, txns[i])
	}

	return nTxns
}
