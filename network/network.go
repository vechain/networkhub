package network

import (
	"encoding/json"
	"time"

	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
)

type Network struct {
	Environment string      `json:"environment"`
	Nodes       []node.Node `json:"nodes"`
	ID          string      `json:"id"`
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
		n.ID = network.ID
		n.Environment = network.Environment
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

// UnmarshalNode function unmarshals JSON data into the appropriate type based on the presence of VIP212
func UnmarshalNode(data []byte) (node.Node, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	if genesisData, ok := raw["genesis"].(map[string]interface{}); ok {
		genesis.HandleAdditionalFields(&genesisData)
	}

	modifiedData, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	nodeType := &node.BaseNode{}
	if err := json.Unmarshal(modifiedData, &nodeType); err != nil {
		return nil, err
	}

	return nodeType, nil
}

func (n *Network) HealthCheck(block uint32, timeout time.Duration) error {
	for _, n := range n.Nodes {
		if err := n.HealthCheck(block, timeout); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON implements custom unmarshalling for Network
func (n *Network) UnmarshalJSON(data []byte) error {
	type Alias Network
	aux := &struct {
		Nodes []json.RawMessage `json:"nodes"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, nodeData := range aux.Nodes {
		nodeObj, err := UnmarshalNode(nodeData)
		if err != nil {
			return err
		}
		n.Nodes = append(n.Nodes, nodeObj)
	}

	return nil
}
