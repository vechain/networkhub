package node

import (
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
	GetAPIAddr() string
	SetAPIAddr(addr string)
	GetAPICORS() string
	SetAPICORS(origins string)
	GetP2PListenPort() int
	SetP2PListenPort(port int)
	GetAdditionalArgs() map[string]string
	SetAdditionalArgs(args map[string]string)
	AddAdditionalArg(key, value string)
	GetVerbosity() int
	GetHTTPAddr() string
	GetFakeExecution() bool
}

type Lifecycle interface {
	Stop() error
	Start() error
}
