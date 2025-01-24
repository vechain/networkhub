package genesis

import (
	"fmt"

	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

type CustomGenesis struct {
	LaunchTime uint64                   `json:"launchTime"`
	GasLimit   uint64                   `json:"gaslimit"`
	ExtraData  string                   `json:"extraData"`
	Accounts   []genesis.Account        `json:"accounts"`
	Authority  []genesis.Authority      `json:"authority"`
	Params     genesis.Params           `json:"params"`
	Executor   genesis.Executor         `json:"executor"`
	ForkConfig *CustomGenesisForkConfig `json:"forkConfig"`
}

type CustomGenesisForkConfig struct {
	thor.ForkConfig
	AdditionalFields map[string]uint32 `json:"additionalFields,omitempty"`
}

// NewCustomGenesisForkConfig creates a new instance of CustomGenesisForkConfig
func NewCustomGenesisForkConfig(baseConfig thor.ForkConfig) *CustomGenesisForkConfig {
	return &CustomGenesisForkConfig{
		ForkConfig:       baseConfig,
		AdditionalFields: make(map[string]uint32),
	}
}

// AddField adds a new field to the AdditionalFields map in CustomGenesisForkConfig
func (cfg *CustomGenesisForkConfig) AddField(key string, value uint32) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	cfg.AdditionalFields[key] = value
	return nil
}

// GetField retrieves the value of a field from the AdditionalFields map
func (cfg *CustomGenesisForkConfig) GetField(key string) (uint32, bool) {
	value, exists := cfg.AdditionalFields[key]
	return value, exists
}

// RemoveField removes a field from the AdditionalFields map
func (cfg *CustomGenesisForkConfig) RemoveField(key string) error {
	if _, exists := cfg.AdditionalFields[key]; !exists {
		return fmt.Errorf("field %s does not exist", key)
	}
	delete(cfg.AdditionalFields, key)
	return nil
}
