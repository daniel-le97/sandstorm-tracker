package rcon

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

// ClientPool manages RCON connections to multiple servers
type ClientPool struct {
	clients map[string]*RconClient
	configs map[string]*ServerConfig
	logger  *slog.Logger
	mu      sync.RWMutex
}

// ServerConfig contains the configuration for an RCON server
type ServerConfig struct {
	Address  string
	Password string
	Timeout  time.Duration
}

// NewClientPool creates a new RCON client pool
func NewClientPool(logger *slog.Logger) *ClientPool {
	return &ClientPool{
		clients: make(map[string]*RconClient),
		configs: make(map[string]*ServerConfig),
		logger:  logger,
	}
}

// AddServer adds a server configuration to the pool
func (p *ClientPool) AddServer(serverID string, config *ServerConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	p.configs[serverID] = config
}

// RemoveServer removes a server from the pool and closes its connection
func (p *ClientPool) RemoveServer(serverID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[serverID]; exists {
		client.Conn.Close()
		delete(p.clients, serverID)
	}

	delete(p.configs, serverID)
}

// GetClient returns an RCON client for the specified server, creating it if needed
func (p *ClientPool) GetClient(serverID string) (*RconClient, error) {
	p.mu.RLock()
	// Return existing client if available
	if client, exists := p.clients[serverID]; exists {
		p.mu.RUnlock()
		return client, nil
	}

	// Get server config (hold minimal lock)
	config, exists := p.configs[serverID]
	if !exists {
		if p.logger != nil {
			p.logger.Error("Server not configured in RCON pool",
				"server", serverID,
				"available_servers", fmt.Sprintf("%v", p.listConfiguredServersLocked()))
		}
		p.mu.RUnlock()
		return nil, fmt.Errorf("no configuration found for server: %s", serverID)
	}

	// Copy config to avoid holding lock during network I/O
	configCopy := *config
	p.mu.RUnlock()

	// Create new client outside of lock (allows other goroutines to access pool)
	client, err := p.createClient(serverID, &configCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to create RCON client for %s: %w", serverID, err)
	}

	// Acquire lock to store client
	p.mu.Lock()
	// Double-check another goroutine didn't create it while we were creating ours
	if existing, exists := p.clients[serverID]; exists {
		p.mu.Unlock()
		client.Conn.Close() // Close our redundant client
		return existing, nil
	}

	p.clients[serverID] = client
	p.mu.Unlock()

	if p.logger != nil {
		p.logger.Info("Created RCON client", "server", serverID, "address", configCopy.Address)
	}

	return client, nil
}

// createClient creates and authenticates a new RCON client
func (p *ClientPool) createClient(serverID string, config *ServerConfig) (*RconClient, error) {
	conn, err := net.DialTimeout("tcp", config.Address, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	rconConfig := DefaultConfig()
	if p.logger != nil {
		rconConfig.Logger = p.logger
	}

	client := NewRconClient(conn, rconConfig)

	if !client.Auth(config.Password) {
		conn.Close()
		return nil, fmt.Errorf("authentication failed")
	}

	return client, nil
}

// SendCommand sends an RCON command to a specific server
func (p *ClientPool) SendCommand(serverID string, command string) (string, error) {
	client, err := p.GetClient(serverID)
	if err != nil {
		return "", err
	}

	response, err := client.Send(command)
	if err != nil {
		// If command fails, remove the client so it gets recreated on next attempt
		p.mu.Lock()
		delete(p.clients, serverID)
		p.mu.Unlock()

		if p.logger != nil {
			p.logger.Error("RCON command failed, removed client from pool",
				"server", serverID,
				"command", command,
				"error", err)
		}

		return "", err
	}

	return response, nil
}

// CloseAll closes all RCON connections in the pool
func (p *ClientPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for serverID, client := range p.clients {
		client.Conn.Close()
		if p.logger != nil {
			p.logger.Info("Closed RCON client", "server", serverID)
		}
	}

	p.clients = make(map[string]*RconClient)
}

// ListServers returns all server IDs in the pool
func (p *ClientPool) ListServers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	servers := make([]string, 0, len(p.configs))
	for serverID := range p.configs {
		servers = append(servers, serverID)
	}

	return servers
}

// listConfiguredServersLocked returns all configured server IDs (must be called with lock held)
func (p *ClientPool) listConfiguredServersLocked() []string {
	servers := make([]string, 0, len(p.configs))
	for serverID := range p.configs {
		servers = append(servers, serverID)
	}
	return servers
}

// IsConnected checks if a client is currently connected for the given server
func (p *ClientPool) IsConnected(serverID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.clients[serverID]
	return exists
}
