package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

type BaseNode struct {
	ID            string `json:"id"` //TODO this is a mandatory field
	Key           string `json:"key"`
	APIAddr       string `json:"apiAddr"`
	APICORS       string `json:"apiCORS"`
	ConfigDir     string `json:"configDir,omitempty"`
	DataDir       string `json:"dataDir,omitempty"`
	ExecArtifact  string `json:"execArtifact"` // used to determine the executing version of the node ( path, dockerImage, etc)
	P2PListenPort int    `json:"p2pListenPort"`
	Verbosity     int    `json:"verbosity"`
	EnodeData     string `json:"enode"` // todo: this should be a generated method
	Type          string `json:"type"`
	FakeExecution bool   `json:"fakeExecution"`
}

func (b *BaseNode) GetVerbosity() int {
	return b.Verbosity
}

func (b *BaseNode) GetP2PListenPort() int {
	return b.P2PListenPort
}

func (b *BaseNode) GetAPIAddr() string {
	return b.APIAddr
}

func (b *BaseNode) GetAPICORS() string {
	return b.APICORS
}

func (b *BaseNode) GetGenesis() any {
	return b.GetExecArtifact()
}

func (b *BaseNode) GetKey() string {
	return b.Key
}

func New() Node {
	return &BaseNode{}
}

func (b *BaseNode) GetConfigDir() string {
	return b.ConfigDir
}

func (b *BaseNode) SetConfigDir(s string) {
	b.ConfigDir = s
}

func (b *BaseNode) GetDataDir() string {
	return b.DataDir
}

func (b *BaseNode) SetDataDir(s string) {
	b.DataDir = s
}

func (b *BaseNode) GetID() string {
	return b.ID
}

func (b *BaseNode) GetExecArtifact() string {
	return b.ExecArtifact
}

func (b *BaseNode) SetExecArtifact(artifact string) {
	b.ExecArtifact = artifact
}

func (b *BaseNode) GetHTTPAddr() string {
	//todo make this smarter
	if strings.Contains(b.APIAddr, "0.0.0.0") {
		return "http://" + strings.ReplaceAll(b.APIAddr, "0.0.0.0", "127.0.0.1")
	}
	return "http://" + b.APIAddr
}

func (b *BaseNode) GetFakeExecution() bool {
	return b.FakeExecution
}

func (b *BaseNode) Enode(ipAddr string) (string, error) {
	privKey, err := crypto.HexToECDSA(b.Key)
	if err != nil {
		return "", fmt.Errorf("unable to process key for node %s : %w", b.ID, err)
	}

	return fmt.Sprintf("enode://%x@%s:%v", discover.PubkeyID(&privKey.PublicKey).Bytes(), ipAddr, b.P2PListenPort), nil
}
