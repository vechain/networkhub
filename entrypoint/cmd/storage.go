package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/vechain/networkhub/network"
	"io/ioutil"
	"log/slog"
	"os"
)

type StorageJson struct {
	Network map[string]*network.Network `json:"network"`
}

type Storage struct {
	path string
}

func (s *Storage) Store(networkID string, net *network.Network) error {
	storageJson, err := s.LoadExistingNetworks()
	if err != nil {
		return fmt.Errorf("unable to load existing networks: %w", err)
	}

	// Add/Update the network entry
	storageJson[networkID] = net

	// Marshal the updated data
	data, err := json.MarshalIndent(storageJson, "", "  ")
	if err != nil {
		return err
	}

	// Write the updated data back to file
	err = ioutil.WriteFile(s.path, data, 0644)
	if err != nil {
		return err
	}

	slog.Info("Network saved to file", "filepath", s.path)
	return nil
}

func (s *Storage) LoadExistingNetworks() (map[string]*network.Network, error) {
	// Initialize an empty StorageJson
	storageJson := make(map[string]*network.Network)

	// Check if file exists
	if _, err := os.Stat(s.path); err == nil {
		// File exists, load the current data
		fileData, err := ioutil.ReadFile(s.path)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(fileData, &storageJson)
		if err != nil {
			return nil, err
		}
	}
	return storageJson, nil
}

func NewStorage(path string) *Storage {
	return &Storage{
		path: path,
	}
}
