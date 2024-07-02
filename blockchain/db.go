package blockchain

import (
	"encoding/json"

	"github.com/sap200/evochain/constants"
	"github.com/syndtr/goleveldb/leveldb"
)

func PutIntoDb(bs BlockchainStruct) error {
	db, err := leveldb.OpenFile(constants.BLOCKCHAIN_DB_PATH, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	// save into the database
	value, err := json.Marshal(bs)
	if err != nil {
		return err
	}
	err = db.Put([]byte(constants.BLOCKCHAIN_KEY), value, nil)
	if err != nil {
		return err
	}

	return nil
}

func GetBlockchain() (*BlockchainStruct, error) {
	db, err := leveldb.OpenFile(constants.BLOCKCHAIN_DB_PATH, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	data, err := db.Get([]byte(constants.BLOCKCHAIN_KEY), nil)
	if err != nil {
		return nil, err
	}

	var bc BlockchainStruct
	err = json.Unmarshal(data, &bc)
	if err != nil {
		return nil, err
	}

	return &bc, nil
}

func KeyExists() (bool, error) {
	db, err := leveldb.OpenFile(constants.BLOCKCHAIN_DB_PATH, nil)
	if err != nil {
		return false, err
	}
	defer db.Close()

	exists, err := db.Has([]byte(constants.BLOCKCHAIN_KEY), nil)
	return exists, err
}
