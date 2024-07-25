package consts

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

var (
	LargeBigValue = new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))
)
