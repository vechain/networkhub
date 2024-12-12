package consts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	LargeBigValue = new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))
)
