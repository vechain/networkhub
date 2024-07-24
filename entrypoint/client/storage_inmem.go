package client

import (
	"github.com/vechain/networkhub/network"
	"log/slog"
)

type Storage struct {
	path    string
	storage map[string]*network.Network
}

func NewInMemStorage() *Storage {
	return &Storage{}
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
