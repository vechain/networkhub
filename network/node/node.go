package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/vechain/thor/v2/genesis"
)

const (
	MasterNode  = "masterNode"
	RegularNode = "regularNode"
)

type Node struct {
	ID            string                 `json:"id"`                //TODO this is a mandatory field
	Genesis       *genesis.CustomGenesis `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
	DataDir       string                 `json:"dataDir,omitempty"`
	ConfigDir     string                 `json:"configDir,omitempty"`
	P2PListenPort int                    `json:"p2pListenPort"`
	APIAddr       string                 `json:"apiAddr"`
	APICORS       string                 `json:"apiCORS"`
	Type          string                 `json:"type"`
	Key           string                 `json:"key"`
	EnodeData     string                 `json:"enode"`        // todo: this should be a generated method
	ExecArtifact  string                 `json:"execArtifact"` // used to determine the executing version of the node ( path, dockerImage, etc)
	Verbosity     int                    `json:"verbosity"`
}

func (n *Node) Enode(ipAddr string) (string, error) {
	privKey, err := crypto.HexToECDSA(n.Key)
	if err != nil {
		return "", fmt.Errorf("unable to process key for node %s : %w", n.ID, err)
	}

	return fmt.Sprintf("enode://%x@%s:%v", discover.PubkeyID(&privKey.PublicKey).Bytes(), ipAddr, n.P2PListenPort), nil
}
