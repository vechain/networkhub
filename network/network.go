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

	// Phase A: Node Health - Check if each node is healthy and can fetch blocks
	if err := n.checkNodeHealth(block, timeout); err != nil {
		return fmt.Errorf("node health check failed: %w", err)
	}

	// Phase B: Peer Connectivity - Check if nodes are properly connected (skip for public networks)
	if err := n.checkPeerConnectivity(timeout); err != nil {
		return fmt.Errorf("peer connectivity check failed: %w", err)
	}

	// Phase C: Block Consistency - Check if all nodes have the same block hash
	if err := n.checkBlockConsistency(block); err != nil {
		return fmt.Errorf("block consistency check failed: %w", err)
	}

	return nil
}

// checkNodeHealth verifies each node can fetch the specified block
func (n *Network) checkNodeHealth(block uint32, timeout time.Duration) error {
	for _, node := range n.Nodes {
		if err := node.HealthCheck(block, timeout); err != nil {
			return fmt.Errorf("node %s health check failed: %w", node.GetID(), err)
		}
	}
	return nil
}

// checkPeerConnectivity verifies all nodes are connected to expected number of peers
func (n *Network) checkPeerConnectivity(timeout time.Duration) error {
	// Skip peer connectivity check for public networks
	if n.hasPublicNetworkNodes() {
		return nil
	}

	expectedPeerCount := len(n.Nodes) - 1
	if expectedPeerCount <= 0 {
		return nil // Single node network or empty - no peer connectivity to check
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(time.Until(deadline)):
			return fmt.Errorf("timeout waiting for peer connectivity - expected %d peers per node", expectedPeerCount)
		case <-ticker.C:
			allConnected := true
			for _, node := range n.Nodes {
				client := thorclient.New(node.GetHTTPAddr())
				peers, err := client.Peers()
				if err != nil {
					return fmt.Errorf("failed to get peers for node %s: %w", node.GetID(), err)
				}
				if len(peers) != expectedPeerCount {
					allConnected = false
					break
				}
			}

			if allConnected {
				return nil
			}

			// Continue polling
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for peer connectivity - expected %d peers per node", expectedPeerCount)
			}
		}
	}
}

// checkBlockConsistency verifies all nodes return the same block hash
func (n *Network) checkBlockConsistency(block uint32) error {
	var baseBlk *api.JSONCollapsedBlock
	for _, node := range n.Nodes {
		client := thorclient.New(node.GetHTTPAddr())
		nodeBlk, err := client.Block(fmt.Sprintf("%d", block))
		if err != nil {
			return fmt.Errorf("failed to get block %d from node %s: %w", block, node.GetID(), err)
		}
		if baseBlk == nil {
			baseBlk = nodeBlk
		} else if baseBlk.ID != nodeBlk.ID {
			return fmt.Errorf(
				"block hash mismatch at height %d - node %s has %s, expected %s",
				block, node.GetID(), nodeBlk.ID.String(), baseBlk.ID.String(),
			)
		}
	}
	return nil
}

// hasPublicNetworkNodes checks if any node is configured for public networks (testnet/mainnet)
func (n *Network) hasPublicNetworkNodes() bool {
	for _, node := range n.Nodes {
		if isPublicNetworkNode(node) {
			return true
		}
	}
	return false
}

// isPublicNetworkNode checks if a node is configured for a public network
func isPublicNetworkNode(node node.Config) bool {
	networkArg, exists := node.GetAdditionalArgs()["network"]
	return exists && (networkArg == "test" || networkArg == "main")
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
