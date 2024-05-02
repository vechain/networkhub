package common

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/vechain/thor/v2/thor"
)

var (
	NotFoundErr     = fmt.Errorf("not found")
	Not200StatusErr = fmt.Errorf("not 200 status code")
)

type Account struct {
	Address    *thor.Address
	PrivateKey *ecdsa.PrivateKey
}

type TxSendResult struct {
	ID *thor.Bytes32
}
