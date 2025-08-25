package common

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
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
	for range maxRetries {
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

type TxSendResult struct {
	ID *thor.Bytes32
}

var getPeerCount = func(httpAddr string) (int, error) {
	peers, err := thorclient.New(httpAddr).Peers()
	if err != nil {
		return 0, err
	}
	return len(peers), nil
}

// WaitForPeersConnection waits until every node sees all other nodes as peers.
func WaitForPeersConnection(nodes []node.Config, ctx context.Context) error {
	if len(nodes) == 0 {
		return nil
	}

	expected := len(nodes) - 1

	ctxWithTimeout := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctxWithTimeout, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	check := func() bool {
		for _, n := range nodes {
			count, err := getPeerCount(n.GetHTTPAddr())
			if err != nil || count < expected {
				return false
			}
		}
		return true
	}

	if check() {
		return nil
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return errors.New("timed out waiting for nodes to connect")
		case <-ticker.C:
			if check() {
				return nil
			}
		}
	}
}
