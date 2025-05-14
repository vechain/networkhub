package node

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/thor/v2/thorclient"
)

type BaseNode struct {
	ID             string                 `json:"id"` //TODO this is a mandatory field
	Key            string                 `json:"key"`
	APIAddr        string                 `json:"apiAddr"`
	APICORS        string                 `json:"apiCORS"`
	ConfigDir      string                 `json:"configDir,omitempty"`
	DataDir        string                 `json:"dataDir,omitempty"`
	ExecArtifact   string                 `json:"execArtifact"` // used to determine the executing version of the node ( path, dockerImage, etc)
	P2PListenPort  int                    `json:"p2pListenPort"`
	Verbosity      int                    `json:"verbosity"`
	EnodeData      string                 `json:"enode"` // todo: this should be a generated method
	Type           string                 `json:"type"`
	FakeExecution  bool                   `json:"fakeExecution"`
	Genesis        *genesis.CustomGenesis `json:"genesis"`
	AdditionalArgs map[string]string      `json:"additionalArgs"`
}

func (b *BaseNode) GetVerbosity() int {
	return b.Verbosity
}

func (b *BaseNode) GetP2PListenPort() int {
	return b.P2PListenPort
}

func (b *BaseNode) SetP2PListenPort(port int) {
	b.P2PListenPort = port
}
func (b *BaseNode) GetAPIHost() string {
	split := strings.Split(b.GetAPIAddr(), ":")
	if len(split) != 2 {
		panic(fmt.Errorf("unable to parse API Addr"))
	}
	return split[0]
}
func (b *BaseNode) SetAPIHost(host string) {
	split := strings.Split(b.GetAPIAddr(), ":")
	if len(split) != 2 {
		panic(fmt.Errorf("unable to parse API Addr"))
	}
	b.APIAddr = fmt.Sprintf("%s:%s", host, split[1])
}
func (b *BaseNode) GetAPIAddr() string {
	return b.APIAddr
}
func (b *BaseNode) SetAPIAddr(addr string) {
	b.APIAddr = addr
}

func (b *BaseNode) GetAPICORS() string {
	return b.APICORS
}

func (b *BaseNode) GetKey() string {
	return b.Key
}

func New() Config {
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

func (b *BaseNode) SetID(id string) {
	b.ID = id
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

func (b *BaseNode) GetAdditionalArgs() map[string]string {
	return b.AdditionalArgs
}

func (b *BaseNode) SetAdditionalArgs(args map[string]string) {
	b.AdditionalArgs = args
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

func (b *BaseNode) HealthCheck(block uint32, timeout time.Duration) error {
	client := thorclient.New(b.GetHTTPAddr())
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return fmt.Errorf("timeout waiting for node %s to be healthy", b.ID)
		default:
			blk, err := client.Block(strconv.Itoa(int(block)))
			if err == nil && blk != nil {
				return nil
			}
			slog.Debug("waiting for node to be healthy", "node", b.ID, "block", block, "error", err)
			time.Sleep(1 * time.Second)
		}
	}
}

func (b *BaseNode) GetGenesis() *genesis.CustomGenesis {
	return b.Genesis
}

func (b *BaseNode) SetGenesis(genesis *genesis.CustomGenesis) {
	b.Genesis = genesis
}
