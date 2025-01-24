package preset

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vechain/thor/v2/genesis"
)

func TestGenesisUnmarshal(t *testing.T) {
	type derp struct {
		Hex *genesis.HexOrDecimal256
	}

	ble := genesis.HexOrDecimal256(*big.NewInt(123))
	t.Log(derp{Hex: &ble})

	marshalJSON, err := ble.MarshalJSON()
	require.NoError(t, err)
	t.Log(string(marshalJSON))

	marshal, err := json.Marshal(ble)
	require.NoError(t, err)
	t.Log(marshal)
}
