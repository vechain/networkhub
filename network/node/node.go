package node

const (
	MasterNode  = "masterNode"
	RegularNode = "regularNode"
)

type Lifecycle interface {
	Stop() error
	Start() error
}
