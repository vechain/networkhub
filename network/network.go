package network

import (
	"encoding/json"
	"fmt"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thorclient"
	"time"

	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/networkhub/thorbuilder"
)

type Network struct {
	Environment string              `json:"environment"`
	Nodes       []node.Config       `json:"nodes"`
	BaseID      string              `json:"baseid"`
	ThorBuilder *thorbuilder.Config `json:"thorBuilder,omitempty"`
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
		n.BaseID = network.BaseID
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
func UnmarshalNode(data []byte) (node.Config, error) {
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
	if len(n.Nodes) == 0 {
		return fmt.Errorf("no nodes defined in the network")
	}

	var baseBlk *api.JSONCollapsedBlock
	for _, n := range n.Nodes {
		if err := n.HealthCheck(block, timeout); err != nil {
			return err
		}
		nodeBlk, err := thorclient.New(n.GetHTTPAddr()).Block(fmt.Sprintf("%d", block))
		if err != nil {
			return err
		}
		if baseBlk == nil {
			baseBlk = nodeBlk
		} else if baseBlk.ID != nodeBlk.ID {
			return fmt.Errorf(
				"unexpected blocks at the same height - node: %s height: %d hashNewBlk: %s hashBlk: %s",
				n.GetID(), block, baseBlk.ID.String(), nodeBlk.ID.String(),
			)
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

func (n *Network) ID() string {
	return n.Environment + n.BaseID
}
