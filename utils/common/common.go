package common

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/thor/v2/thor"
)

var (
	NotFoundErr     = fmt.Errorf("not found")
	Not200StatusErr = fmt.Errorf("not 200 status code")
)

func NewAccount(pkString string) *Account {
	pk, err := crypto.HexToECDSA(pkString)
	if err != nil {
		panic(err)
	}
	addr := thor.Address(crypto.PubkeyToAddress(pk.PublicKey))
	return &Account{
		Address:    &addr,
		PrivateKey: pk,
	}
}

type Account struct {
	Address    *thor.Address
	PrivateKey *ecdsa.PrivateKey
}

type TxSendResult struct {
	ID *thor.Bytes32
}
