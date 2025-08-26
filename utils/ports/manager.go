package ports

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"sync"
)

// Manager is a simple in-memory port allocator.
// It tracks allocated ports per networkID and ensures they are available
// by binding before returning them.
type Manager struct {
	mu        sync.Mutex
	ports     map[int]bool     // port -> used
	portsByID map[string][]int // id -> ports
}

var defaultManager = &Manager{
	ports:     make(map[int]bool),
	portsByID: make(map[string][]int),
}

// Default returns the process-wide manager.
func Default() *Manager { return defaultManager }

// Allocate returns an available TCP port and records it under networkID.
func (m *Manager) Allocate(networkID string) (int, error) {
	if networkID == "" {
		return 0, errors.New("networkID must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.portsByID[networkID]; !exists {
		m.portsByID[networkID] = make([]int, 0)
	}

	port, err := m.findFreePort()
	if err != nil {
		return 0, err
	}
	m.ports[port] = true
	m.portsByID[networkID] = append(m.portsByID[networkID], port)
	return port, nil
}

// ReleaseAll frees all ports reserved for the given networkID.
func (m *Manager) ReleaseAll(networkID string) error {
	if networkID == "" {
		return errors.New("networkID must not be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	ports, ok := m.portsByID[networkID]
	if ok {
		for _, p := range ports {
			delete(m.ports, p)
		}
		delete(m.portsByID, networkID)
	}
	return nil
}

// findFreePort returns a free TCP port in the ephemeral range, avoiding collisions
// recorded in memory. It validates availability by binding and closing.
func (m *Manager) findFreePort() (int, error) {
	const (
		minPort     = 49152
		maxPort     = 65535
		maxAttempts = 100
	)

	// Try random sampling first
	for range maxAttempts {
		port := randomPortInRange(minPort, maxPort)
		if m.ports[port] {
			continue
		}
		if isPortAvailable(port) {
			return port, nil
		}
	}

	// Fallback to sequential scan
	for port := minPort; port <= maxPort; port++ {
		if m.ports[port] {
			continue
		}
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, errors.New("no free ports available in ephemeral range")
}

func randomPortInRange(minPort, maxPort int) int {
	var b [2]byte
	_, _ = rand.Read(b[:])
	n := int(b[0])<<8 | int(b[1])
	return minPort + (n % (maxPort - minPort + 1))
}

// isPortAvailable checks if a port is actually available for binding.
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
