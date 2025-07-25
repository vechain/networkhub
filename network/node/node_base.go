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

type Config struct {
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

func (b *Config) GetVerbosity() int {
	return b.Verbosity
}

func (b *Config) GetP2PListenPort() int {
	return b.P2PListenPort
}

func (b *Config) SetP2PListenPort(port int) {
	b.P2PListenPort = port
}
func (b *Config) GetAPIHost() string {
	split := strings.Split(b.GetAPIAddr(), ":")
	if len(split) != 2 {
		panic(fmt.Errorf("unable to parse API Addr"))
	}
	return split[0]
}
func (b *Config) SetAPIHost(host string) {
	split := strings.Split(b.GetAPIAddr(), ":")
	if len(split) != 2 {
		panic(fmt.Errorf("unable to parse API Addr"))
	}
	b.APIAddr = fmt.Sprintf("%s:%s", host, split[1])
}
func (b *Config) GetAPIAddr() string {
	return b.APIAddr
}
func (b *Config) SetAPIAddr(addr string) {
	b.APIAddr = addr
}

func (b *Config) GetAPICORS() string {
	return b.APICORS
}

func (b *Config) GetKey() string {
	return b.Key
}

func New() *Config {
	return &Config{}
}

func (b *Config) GetConfigDir() string {
	return b.ConfigDir
}

func (b *Config) SetConfigDir(s string) {
	b.ConfigDir = s
}

func (b *Config) GetDataDir() string {
	return b.DataDir
}

func (b *Config) SetDataDir(s string) {
	b.DataDir = s
}

func (b *Config) SetID(id string) {
	b.ID = id
}

func (b *Config) GetID() string {
	return b.ID
}

func (b *Config) GetExecArtifact() string {
	return b.ExecArtifact
}

func (b *Config) SetExecArtifact(artifact string) {
	b.ExecArtifact = artifact
}

func (b *Config) GetAdditionalArgs() map[string]string {
	return b.AdditionalArgs
}

func (b *Config) SetAdditionalArgs(args map[string]string) {
	b.AdditionalArgs = args
}

func (b *Config) GetHTTPAddr() string {
	//todo make this smarter
	if strings.Contains(b.APIAddr, "0.0.0.0") {
		return "http://" + strings.ReplaceAll(b.APIAddr, "0.0.0.0", "127.0.0.1")
	}
	return "http://" + b.APIAddr
}

func (b *Config) GetFakeExecution() bool {
	return b.FakeExecution
}

func (b *Config) Enode(ipAddr string) (string, error) {
	privKey, err := crypto.HexToECDSA(b.Key)
	if err != nil {
		return "", fmt.Errorf("unable to process key for node %s : %w", b.ID, err)
	}

	return fmt.Sprintf("enode://%x@%s:%v", discover.PubkeyID(&privKey.PublicKey).Bytes(), ipAddr, b.P2PListenPort), nil
}

func (b *Config) HealthCheck(block uint32, timeout time.Duration) error {
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

func (b *Config) GetGenesis() *genesis.CustomGenesis {
	return b.Genesis
}

func (b *Config) SetGenesis(genesis *genesis.CustomGenesis) {
	b.Genesis = genesis
}
