package network

import (
	"encoding/json"
	"fmt"

	"github.com/vechain/networkhub/network/node"
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
		if forkConfig, ok := genesisData["forkConfig"].(map[string]interface{}); ok {
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
				genesisData["forkConfig"] = forkConfig
			}
		}
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

	fmt.Printf("Numero de nodos %v\n", n.Nodes[0].GetGenesis().ForkConfig.ForkConfig)

	return nil
}
