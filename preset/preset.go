package preset

import (
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/vechain/networkhub/network"
	"github.com/vechain/thor/v2/genesis"
)

type APIConfigPayload struct {
	ArtifactPath string `json:"artifactPath"`
}

type Networks struct {
	presets map[string]*network.Network
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func NewPresetNetworks() *Networks {
	return &Networks{
		presets: map[string]*network.Network{},
	}
}

func (p *Networks) Register(id string, preset *network.Network) {
	p.presets[id] = preset
	slog.Info("Registered preset network", "networkId", id)
}

func (p *Networks) Load(id string, configPayload *APIConfigPayload) (*network.Network, error) {
	preset, ok := p.presets[id]
	if !ok {
		return nil, fmt.Errorf("unable to find preset with id %s", id)
	}

	if configPayload == nil || configPayload.ArtifactPath == "" {
		return nil, fmt.Errorf("preset config must be set")
	}

	if !fileExists(configPayload.ArtifactPath) {
		return nil, fmt.Errorf("file does not exist at location: %s", configPayload.ArtifactPath)
	}

	// override the default path
	for _, node := range preset.Nodes {
		node.ExecArtifact = configPayload.ArtifactPath
	}
	return preset, nil
}

func convToHexOrDecimal256(i *big.Int) *genesis.HexOrDecimal256 {
	tmp := genesis.HexOrDecimal256(*i)
	return &tmp
}
