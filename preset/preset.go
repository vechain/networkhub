package preset

import (
	"fmt"
	"log/slog"
	"math/big"

	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/thor/v2/genesis"
)

var (
	Account1 = common.NewAccount("b2c859e115ef4a3f5e4d32228b41de4c661c527a32f723ac37745bf860fd09cb") // 0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e
	Account2 = common.NewAccount("4de650ca1c8beae4ed6a4358087f50c01b51f5c0002ae9836c55039ca9818d0c") // 0x5c29518F6a6124a2BeE89253347c8295f604710A
	Account3 = common.NewAccount("1b310ea04afd6d14a8f142158873fc70bfd4ba12a19138cc5b309fce7c77105e") // 0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE
	Account4 = common.NewAccount("c70dda88e779df10abbc7c5d37fbb3478c5cf8df2a70d6b0bfc551a5a9a17359") // 0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c
	Account5 = common.NewAccount("ade54b623a4f4afc38f962a85df07a428204a67cee0c9b43a99ca255fd2fb9a6") // 0x0aeC31606e217895696771961de416Efa185Be66
	Account6 = common.NewAccount("92ad65923d6782a43e6a1be01a8e52bce701967d78937e73da746a58f293ba30") // 0x9C2871C411CCe579B987E9b932C484dA8b901075
)

type APIConfigPayload struct {
	ArtifactPath string `json:"artifactPath"`
	Environment  string `json:"environment"`
}

type Networks struct {
	presets map[string]*network.Network
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

	// override the default path
	for _, node := range preset.Nodes {
		node.SetExecArtifact(configPayload.ArtifactPath)
	}
	// override the default environment
	preset.Environment = configPayload.Environment
	return preset, nil
}

func convToHexOrDecimal256(i *big.Int) *genesis.HexOrDecimal256 {
	tmp := genesis.HexOrDecimal256(*i)
	return &tmp
}
