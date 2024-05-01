package environments

import "github.com/vechain/networkhub/network"

type Actions interface {
	LoadConfig(cfg *network.Network) (string, error)
	StartNetwork() error
	StopNetwork() error
	Info() error
}

const (
	Local = "local"
)
