package genesis

import (
	"encoding/json"

	thorgenesis "github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

// CustomGenesis wraps Thor's CustomGenesis, promoting all its fields (including Stakers)
// to the top level. ForkConfig and Config shadow the embedded struct's fields so the
// local Config type is preserved.
type CustomGenesis struct {
	*thorgenesis.CustomGenesis
	ForkConfig *CustomGenesisForkConfig `json:"forkConfig"`
	Config     *Config                  `json:"config,omitempty"`
}

func Marshal(customGenesis *CustomGenesis) ([]byte, error) {
	return json.Marshal(customGenesis)
}

type CustomGenesisForkConfig struct {
	thor.ForkConfig
}
