package common

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/thor/v2/thor"
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

// Retry retries the given function fn until it succeeds or the maximum number of retries is reached.
// It waits for retryPeriod between each retry.
func Retry(fn func() error, retryPeriod time.Duration, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(retryPeriod)
	}
	return err
}

type Account struct {
	Address    *thor.Address
	PrivateKey *ecdsa.PrivateKey
}

func (a Account) PrivateKeyString() string {
	return fmt.Sprintf("%x", a.PrivateKey.D.Bytes())
}

type TxSendResult struct {
	ID *thor.Bytes32
}
