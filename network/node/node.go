package node

import (
	"time"

	"github.com/vechain/networkhub/network/node/genesis"
)

const (
	MasterNode  = "masterNode"
	RegularNode = "regularNode"
)

type Config interface {
	Enode(ipAddr string) (string, error)
	SetExecArtifact(artifact string)
	GetConfigDir() string
	SetConfigDir(join string)
	GetDataDir() string
	SetDataDir(join string)
	SetID(id string)
	GetID() string
	GetExecArtifact() string
	GetKey() string
	GetGenesis() *genesis.CustomGenesis
	SetGenesis(genesis *genesis.CustomGenesis)
	SetAPIHost(host string)
	GetAPIHost() string
	GetAPIAddr() string
	SetAPIAddr(addr string)
	GetAPICORS() string
	GetP2PListenPort() int
	SetP2PListenPort(port int)
	GetAdditionalArgs() map[string]string
	SetAdditionalArgs(args map[string]string)
	AddAdditionalArg(key, value string)
	GetVerbosity() int
	GetHTTPAddr() string
	GetFakeExecution() bool
	HealthCheck(block uint32, timeout time.Duration) error
}

type Lifecycle interface {
	Stop() error
	Start() error
}
