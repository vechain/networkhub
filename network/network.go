package network

import (
	"encoding/json"

	"github.com/vechain/networkhub/network/node"
)

type Network struct {
	Environment string       `json:"environment"`
	Nodes       []*node.Node `json:"nodes"`
	ID          string       `json:"id"`
}
type Builder struct {
}

type BuilderOptionsFunc func(*Network) error

func WithJSON(s string) BuilderOptionsFunc {
	return func(n *Network) error {
		var network Network // todo: yuck
		err := json.Unmarshal([]byte(s), &network)
		if err != nil {
			return err
		}
		n.Nodes = network.Nodes
		return nil
	}
}

func NewNetwork(opts ...BuilderOptionsFunc) (*Network, error) {
	n := &Network{}
	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
	}
	return n, nil
}
