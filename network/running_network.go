package network

type RunningNode struct{}

type RunningNetwork struct {
	nodes []*RunningNode
}

func (n *RunningNetwork) AddNode(node *RunningNode) {
	n.nodes = append(n.nodes, node)
}
