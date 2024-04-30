package datagen

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/crypto"
	common2 "github.com/vechain/networkhub/utils/common"
	"github.com/vechain/thor/thor"
)

func RandAccount() *common2.Account {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err) // should never happen
	}
	addr := thor.Address(crypto.PubkeyToAddress(key.PublicKey))
	return &common2.Account{
		Address:    &addr,
		PrivateKey: key,
	}
}

func RandAddress() (addr thor.Address) {
	rand.Read(addr[:])
	return
}

func RandKey() (key thor.Bytes32) {
	rand.Read(key[:])
	return
}

func RandUInt64() uint64 {
	var num uint64
	binary.Read(rand.Reader, binary.BigEndian, &num)
	return num
}
