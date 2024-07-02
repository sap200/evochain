package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/sap200/evochain/blockchain"
	"github.com/sap200/evochain/constants"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey `json:"private_key"`
	PublicKey  *ecdsa.PublicKey  `json:"public_key"`
}

func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	wallet := new(Wallet)
	wallet.PrivateKey = privateKey
	wallet.PublicKey = &privateKey.PublicKey

	return wallet, nil
}

func NewWalletFromPrivateKeyHex(privateKeyHex string) *Wallet {
	pk := privateKeyHex[2:]
	d := new(big.Int)
	d.SetString(pk, 16)

	var npk ecdsa.PrivateKey
	npk.D = d
	npk.PublicKey.Curve = elliptic.P256()
	npk.PublicKey.X, npk.PublicKey.Y = npk.PublicKey.Curve.ScalarBaseMult(d.Bytes())

	wallet := new(Wallet)
	wallet.PrivateKey = &npk
	wallet.PublicKey = &npk.PublicKey

	return wallet
}

func (w *Wallet) GetPrivateKeyHex() string {
	return fmt.Sprintf("0x%x", w.PrivateKey.D)
}

func (w *Wallet) GetPublicKeyHex() string {
	return fmt.Sprintf("0x%x%x", w.PublicKey.X, w.PublicKey.Y)
}

func (w *Wallet) GetAddress() string {
	hash := sha256.Sum256([]byte(w.GetPublicKeyHex()[2:]))
	hex := fmt.Sprintf("%x", hash[:])
	address := constants.ADDRESS_PREFIX + hex[len(hex)-40:]
	return address
}

func (w *Wallet) GetSignedTxn(unsignedTxn blockchain.Transaction) (*blockchain.Transaction, error) {
	bs, err := json.Marshal(unsignedTxn)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(bs)

	sig, err := ecdsa.SignASN1(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	var signedTxn blockchain.Transaction
	signedTxn.From = unsignedTxn.From
	signedTxn.To = unsignedTxn.To
	signedTxn.Data = unsignedTxn.Data
	signedTxn.Status = unsignedTxn.Status
	signedTxn.Value = unsignedTxn.Value
	signedTxn.Timestamp = unsignedTxn.Timestamp
	signedTxn.TransactionHash = unsignedTxn.TransactionHash
	// new fields
	signedTxn.Signature = sig
	signedTxn.PublicKey = w.GetPublicKeyHex()

	return &signedTxn, nil
}
