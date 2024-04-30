package node

type Node struct {
	ID            string `json:"id"`
	Genesis       string `json:"genesis,omitempty"`
	P2PListenPort int    `json:"p2pListenPort"`
	DataDir       string `json:"dataDir"`
	ConfigDir     string `json:"configDir"`
	APIAddr       string `json:"apiAddr"`
	APICORS       string `json:"apiCORS"`
	Type          string `json:"type"`
	Key           string `json:"key"`
	Enode         string `json:"enode"`
}
