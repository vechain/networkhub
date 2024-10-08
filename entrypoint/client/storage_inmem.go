package client

import (
	"log/slog"

	"github.com/vechain/networkhub/network"
)

type Storage struct {
	storage map[string]*network.Network
}

func NewInMemStorage() *Storage {
	return &Storage{
		storage: map[string]*network.Network{},
	}
}

func (s *Storage) Store(networkID string, net *network.Network) error {
	// Add/Update the network entry
	s.storage[networkID] = net

	slog.Info("Network saved to memory")
	return nil
}

func (s *Storage) LoadExistingNetworks() (map[string]*network.Network, error) {
	return s.storage, nil
}
