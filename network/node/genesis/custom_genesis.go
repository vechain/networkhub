package genesis

import (
	"encoding/json"
	"fmt"

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

func HandleAdditionalFields(raw *map[string]interface{}) {
	if forkConfig, ok := (*raw)["forkConfig"].(map[string]interface{}); ok {
		// Handle AdditionalFields
		if additionalFields, ok := forkConfig["additionalFields"].(map[string]interface{}); ok {
			for key, value := range additionalFields {
				if num, ok := value.(float64); ok { // JSON numbers are float64 by default
					forkConfig[key] = uint32(num)
					delete(additionalFields, key)
				}
				if len(additionalFields) == 0 {
					delete(forkConfig, "additionalFields")
				}
			}
			(*raw)["forkConfig"] = forkConfig
		}
	}
}

func Marshal(customGenesis *CustomGenesis) ([]byte, error) {
	return json.Marshal(customGenesis)
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
	if cfg.AdditionalFields == nil {
		cfg.AdditionalFields = make(map[string]uint32)
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
