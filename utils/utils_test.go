package utils

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

func TestReadPK(t *testing.T) {
	privateKeyA, err := crypto.HexToECDSA("c55455943bf026dc44fcf189e8765eb0587c94e66029d580bae795386c0b737a")
	if err != nil {
		t.Error(err)
	}
	pubKeyA := privateKeyA.PublicKey
	addrA := crypto.PubkeyToAddress(pubKeyA)
	// Logging this to make sure if the test fails we have the culprit keys
	t.Logf("PK String: %x\n", privateKeyA.D.Bytes())
	t.Logf("PK Address: %s", addrA.Hex())
}

func TestGenerateData(t *testing.T) {
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		t.Error(err)
	}
	pubKeyA := privateKeyA.PublicKey
	addrA := crypto.PubkeyToAddress(pubKeyA)
	// Logging this to make sure if the test fails we have the culprit keys
	t.Logf("PK String: %x\n", privateKeyA.D.Bytes())
	t.Logf("PK Address: %s", addrA.Hex())
}

func TestNodeID(t *testing.T) {
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		t.Error(err)
	}

	pubKeyA := privateKeyA.PublicKey
	addrA := crypto.PubkeyToAddress(pubKeyA)
	// Logging this to make sure if the test fails we have the culprit keys
	t.Logf("PK String: %x\n", privateKeyA.D.Bytes())
	t.Logf("PK Address: %s", addrA.Hex())

	t.Logf("eNode: %s", fmt.Sprintf("enode://%x@[extip]:%v", discover.PubkeyID(&privateKeyA.PublicKey).Bytes(), 8080))
}
