package node

import "github.com/vechain/networkhub/network/node/genesis"

const (
	MasterNode  = "masterNode"
	RegularNode = "regularNode"
)

type Node interface {
	Enode(ipAddr string) (string, error)
	SetExecArtifact(artifact string)
	GetConfigDir() string
	SetConfigDir(join string)
	GetDataDir() string
	SetDataDir(join string)
	GetID() string
	GetExecArtifact() string
	GetKey() string
	GetGenesis() *genesis.CustomGenesis
	SetGenesis(genesis *genesis.CustomGenesis)
	GetAPIAddr() string
	SetAPIAddr(addr string)
	GetAPICORS() string
	GetP2PListenPort() int
	SetP2PListenPort(port int)
	GetVerbosity() int
	GetHTTPAddr() string
	GetFakeExecution() bool
}
