package environments

import "github.com/vechain/networkhub/network"

type Actions interface {
	LoadConfig() error
	StartNetwork(cfg *network.Network) error
	StopNetwork() error
	Info() error
}
