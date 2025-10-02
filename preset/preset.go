package preset

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vechain/thor/v2/genesis"
)

var (
	LargeBigValue = new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))
)

func convToHexOrDecimal256(i *big.Int) *genesis.HexOrDecimal256 {
	tmp := genesis.HexOrDecimal256(*i)
	return &tmp
}

func privateKeyString(k *ecdsa.PrivateKey) string {
	return hexutil.Encode(k.D.Bytes())
}
