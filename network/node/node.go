package node

type Node struct {
	ID            string      `json:"id"`                //TODO this is a mandatory field
	Genesis       interface{} `json:"genesis,omitempty"` //TODO would be nice to have validation in this format
	DataDir       string      `json:"dataDir,omitempty"`
	ConfigDir     string      `json:"configDir,omitempty"`
	P2PListenPort int         `json:"p2pListenPort"`
	APIAddr       string      `json:"apiAddr"`
	APICORS       string      `json:"apiCORS"`
	Type          string      `json:"type"`
	Key           string      `json:"key"`
	Enode         string      `json:"enode"`
}
