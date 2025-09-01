package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/vechain/networkhub/network/node/genesis"
)

type BaseNode struct {
	ID             string                 `json:"id"`
	Key            string                 `json:"key"`
	APIAddr        string                 `json:"apiAddr"`
	APICORS        string                 `json:"apiCORS"`
	ConfigDir      string                 `json:"configDir,omitempty"`
	DataDir        string                 `json:"dataDir,omitempty"`
	ExecArtifact   string                 `json:"execArtifact"`
	P2PListenPort  int                    `json:"p2pListenPort"`
	Verbosity      int                    `json:"verbosity"`
	Type           string                 `json:"type"`
	FakeExecution  bool                   `json:"fakeExecution"`
	Genesis        *genesis.CustomGenesis `json:"genesis"`
	AdditionalArgs map[string]string      `json:"additionalArgs"`
}

func (b *BaseNode) GetVerbosity() int         { return b.Verbosity }
func (b *BaseNode) GetP2PListenPort() int     { return b.P2PListenPort }
func (b *BaseNode) SetP2PListenPort(port int) { b.P2PListenPort = port }

func (b *BaseNode) GetAPIAddr() string     { return b.APIAddr }
func (b *BaseNode) SetAPIAddr(addr string) { b.APIAddr = addr }

func (b *BaseNode) GetAPICORS() string {
	if b.APICORS == "" {
		return "*"
	}
	return b.APICORS
}

func (b *BaseNode) SetAPICORS(origins string) { b.APICORS = origins }

func (b *BaseNode) GetKey() string { return b.Key }

func New() Config { return &BaseNode{} }

func (b *BaseNode) GetConfigDir() string            { return b.ConfigDir }
func (b *BaseNode) SetConfigDir(s string)           { b.ConfigDir = s }
func (b *BaseNode) GetDataDir() string              { return b.DataDir }
func (b *BaseNode) SetDataDir(s string)             { b.DataDir = s }
func (b *BaseNode) SetID(id string)                 { b.ID = id }
func (b *BaseNode) GetID() string                   { return b.ID }
func (b *BaseNode) GetExecArtifact() string         { return b.ExecArtifact }
func (b *BaseNode) SetExecArtifact(artifact string) { b.ExecArtifact = artifact }

func (b *BaseNode) GetAdditionalArgs() map[string]string     { return b.AdditionalArgs }
func (b *BaseNode) SetAdditionalArgs(args map[string]string) { b.AdditionalArgs = args }
func (b *BaseNode) AddAdditionalArg(key, value string) {
	if b.AdditionalArgs == nil {
		b.AdditionalArgs = make(map[string]string)
	}
	b.AdditionalArgs[key] = value
}

func (b *BaseNode) GetHTTPAddr() string {
	if strings.Contains(b.APIAddr, "0.0.0.0") {
		return "http://" + strings.ReplaceAll(b.APIAddr, "0.0.0.0", "127.0.0.1")
	}
	return "http://" + b.APIAddr
}

func (b *BaseNode) GetFakeExecution() bool { return b.FakeExecution }

func (b *BaseNode) Enode(ipAddr string) (string, error) {
	privKey, err := crypto.HexToECDSA(b.Key)
	if err != nil {
		return "", fmt.Errorf("unable to process key for node %s : %w", b.ID, err)
	}
	return fmt.Sprintf("enode://%x@%s:%v", discover.PubkeyID(&privKey.PublicKey).Bytes(), ipAddr, b.P2PListenPort), nil
}

func (b *BaseNode) GetGenesis() *genesis.CustomGenesis        { return b.Genesis }
func (b *BaseNode) SetGenesis(genesis *genesis.CustomGenesis) { b.Genesis = genesis }
