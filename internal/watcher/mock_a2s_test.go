package watcher

import (
	"context"
	"fmt"
	"sync"

	"sandstorm-tracker/internal/a2s"
)

// MockA2SPool is a test implementation of A2S pool that returns predefined responses
type MockA2SPool struct {
	servers map[string]*a2s.ServerStatus
	mu      sync.RWMutex
}

// NewMockA2SPool creates a new mock A2S pool
func NewMockA2SPool() *MockA2SPool {
	return &MockA2SPool{
		servers: make(map[string]*a2s.ServerStatus),
	}
}

// SetServerStatus sets the mock response for a server address
func (m *MockA2SPool) SetServerStatus(address string, status *a2s.ServerStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[address] = status
}

// SetServerOnline is a helper to set a server as online with a specific map
func (m *MockA2SPool) SetServerOnline(address, serverName, mapName string) {
	m.SetServerStatus(address, &a2s.ServerStatus{
		Address: address,
		Online:  true,
		Info: &a2s.ServerInfo{
			Name:       serverName,
			Map:        mapName,
			Players:    0,
			MaxPlayers: 32,
		},
	})
}

// SetServerOffline is a helper to set a server as offline
func (m *MockA2SPool) SetServerOffline(address string) {
	m.SetServerStatus(address, &a2s.ServerStatus{
		Address: address,
		Online:  false,
		Error:   fmt.Errorf("server offline"),
	})
}

// QueryServer returns the mocked server status
func (m *MockA2SPool) QueryServer(ctx context.Context, address string) (*a2s.ServerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.servers[address]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", address)
	}

	if status.Error != nil {
		return status, status.Error
	}

	return status, nil
}

// AddServer is a no-op for the mock
func (m *MockA2SPool) AddServer(address, name string) {
	// Mock doesn't need to track this
}

// RemoveServer is a no-op for the mock
func (m *MockA2SPool) RemoveServer(address string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.servers, address)
}
