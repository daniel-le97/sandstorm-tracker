package a2s

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ServerPool manages queries to multiple servers efficiently
type ServerPool struct {
	client      *Client
	servers     map[string]*Server
	mu          sync.RWMutex
	rateLimiter *RateLimiter
}

// Server represents a monitored server
type Server struct {
	Address   string
	Name      string
	lastInfo  *ServerInfo
	lastError error
	lastQuery time.Time
	mu        sync.RWMutex
}

// ServerStatus contains the current status of a server
type ServerStatus struct {
	Address   string
	Online    bool
	Info      *ServerInfo
	Players   []Player
	Error     error
	QueryTime time.Duration
	LastQuery time.Time
}

// RateLimiter limits queries per server
type RateLimiter struct {
	minInterval time.Duration
	limiters    map[string]*serverLimiter
	mu          sync.Mutex
}

type serverLimiter struct {
	lastQuery time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
		limiters:    make(map[string]*serverLimiter),
	}
}

// Wait blocks until enough time has passed since last query to this server
func (r *RateLimiter) Wait(address string) {
	r.mu.Lock()
	limiter, exists := r.limiters[address]
	if !exists {
		limiter = &serverLimiter{}
		r.limiters[address] = limiter
	}
	r.mu.Unlock()

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	elapsed := time.Since(limiter.lastQuery)
	if elapsed < r.minInterval {
		time.Sleep(r.minInterval - elapsed)
	}
	limiter.lastQuery = time.Now()
}

// NewServerPool creates a new server pool
func NewServerPool() *ServerPool {
	return &ServerPool{
		client:      NewClient(),
		servers:     make(map[string]*Server),
		rateLimiter: NewRateLimiter(1 * time.Second), // 1 query/sec per server
	}
}

// NewServerPoolWithClient creates a server pool with a custom client
func NewServerPoolWithClient(client *Client) *ServerPool {
	return &ServerPool{
		client:      client,
		servers:     make(map[string]*Server),
		rateLimiter: NewRateLimiter(1 * time.Second),
	}
}

// AddServer adds a server to the pool
func (p *ServerPool) AddServer(address string, name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.servers[address] = &Server{
		Address: address,
		Name:    name,
	}
}

// RemoveServer removes a server from the pool
func (p *ServerPool) RemoveServer(address string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.servers, address)
}

// GetServer returns a server by address
func (p *ServerPool) GetServer(address string) (*Server, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	server, exists := p.servers[address]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", address)
	}

	return server, nil
}

// ListServers returns all servers in the pool
func (p *ServerPool) ListServers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	addresses := make([]string, 0, len(p.servers))
	for addr := range p.servers {
		addresses = append(addresses, addr)
	}

	return addresses
}

// QueryServer queries a specific server and returns its status
func (p *ServerPool) QueryServer(ctx context.Context, address string) (*ServerStatus, error) {
	server, err := p.GetServer(address)
	if err != nil {
		return nil, err
	}

	return p.queryServer(ctx, server)
}

// QueryAll queries all servers in the pool concurrently
func (p *ServerPool) QueryAll(ctx context.Context) map[string]*ServerStatus {
	p.mu.RLock()
	servers := make([]*Server, 0, len(p.servers))
	for _, server := range p.servers {
		servers = append(servers, server)
	}
	p.mu.RUnlock()

	results := make(map[string]*ServerStatus)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()

			status, err := p.queryServer(ctx, srv)
			if err != nil {
				status = &ServerStatus{
					Address: srv.Address,
					Online:  false,
					Error:   err,
				}
			}

			mu.Lock()
			results[srv.Address] = status
			mu.Unlock()
		}(server)
	}

	wg.Wait()
	return results
}

// queryServer performs the actual query with rate limiting
func (p *ServerPool) queryServer(ctx context.Context, server *Server) (*ServerStatus, error) {
	start := time.Now()

	// Rate limit per server
	p.rateLimiter.Wait(server.Address)

	// Check context before querying
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Query server info
	info, err := p.client.QueryInfoContext(ctx, server.Address)

	status := &ServerStatus{
		Address:   server.Address,
		QueryTime: time.Since(start),
		LastQuery: time.Now(),
	}

	if err != nil {
		status.Online = false
		status.Error = err
		server.updateStatus(nil, err)
		return status, err
	}

	status.Online = true
	status.Info = info

	// Always query players - Insurgency: Sandstorm may not report player count correctly in info
	// We'll get an empty list if there are no players, which is fine
	players, err := p.client.QueryPlayersContext(ctx, server.Address)
	if err == nil {
		status.Players = players
	} else {
		// Log player query failures for debugging
		fmt.Printf("[A2S] Failed to query players for %s: %v\n", server.Address, err)
	}

	server.updateStatus(info, nil)
	return status, nil
}

// updateStatus updates the server's cached status
func (s *Server) updateStatus(info *ServerInfo, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastInfo = info
	s.lastError = err
	s.lastQuery = time.Now()
}

// GetLastInfo returns the last successful server info query
func (s *Server) GetLastInfo() (*ServerInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastError != nil {
		return nil, s.lastError
	}

	return s.lastInfo, nil
}

// IsOnline returns whether the server was online during last query
func (s *Server) IsOnline() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastError == nil && s.lastInfo != nil
}

// Monitor starts continuous monitoring of all servers
func (p *ServerPool) Monitor(ctx context.Context, interval time.Duration, callback func(map[string]*ServerStatus)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial query
	results := p.QueryAll(ctx)
	callback(results)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results := p.QueryAll(ctx)
			callback(results)
		}
	}
}
