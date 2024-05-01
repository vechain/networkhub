package preset

import (
	"fmt"
	"github.com/vechain/networkhub/network"
)

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
}

func (p *Networks) Load(id string) (*network.Network, error) {
	preset, ok := p.presets[id]
	if !ok {
		return nil, fmt.Errorf("unable to find preset with id %s", id)
	}
	return preset, nil
}
