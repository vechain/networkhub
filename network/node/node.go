package node

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
	GetGenesis() any
	GetAPIAddr() string
	GetAPICORS() string
	GetP2PListenPort() int
	GetVerbosity() int
	GetHTTPAddr() string
	GetFakeExecution() bool
}
