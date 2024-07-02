package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/sap200/evochain/constants"
)

type Block struct {
	BlockNumber  uint64         `json:"block_number"`
	PrevHash     string         `json:"prevHash"`
	Timestamp    int64          `json:"timestamp"`
	Nonce        int            `json:"nonce"`
	Transactions []*Transaction `json:"transactions"`
}

func NewBlock(prevHash string, nonce int, blockNumber uint64) *Block {
	block := new(Block)
	block.PrevHash = prevHash
	block.Timestamp = time.Now().UnixNano()
	block.Nonce = nonce
	block.Transactions = []*Transaction{}
	block.BlockNumber = blockNumber

	return block
}

func (b Block) ToJson() string {
	nb, err := json.Marshal(b)

	if err != nil {
		return err.Error()
	} else {
		return string(nb)
	}
}

func (b Block) Hash() string {

	bs, _ := json.Marshal(b)
	sum := sha256.Sum256(bs)
	hexRep := hex.EncodeToString(sum[:32])
	formattedHexRep := constants.HEX_PREFIX + hexRep

	return formattedHexRep
}

func (b *Block) AddTransactionToTheBlock(txn *Transaction) {
	// check if the txn verification is a success or a failure
	if txn.Status == constants.TXN_VERIFICATION_SUCCESS {
		txn.Status = constants.SUCCESS
	} else {
		txn.Status = constants.FAILED
	}

	b.Transactions = append(b.Transactions, txn)
}
