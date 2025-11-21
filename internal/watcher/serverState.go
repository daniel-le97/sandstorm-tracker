package watcher

import (
	"log/slog"
	"sync"
	"time"
)

// ServerStateTracker manages server activity state and monitoring
type ServerStateTracker struct {
	logger           *slog.Logger
	activeServers    map[string]bool
	activeServersMu  sync.RWMutex
	lastActivity     map[string]time.Time
	lastActivityMu   sync.RWMutex
	onServerActive   func(serverID string)
	onServerInactive func(serverID string)
	callbacksMu      sync.RWMutex
	inactivityTimer  time.Duration
}

// NewServerStateTracker creates a new server state tracker
func NewServerStateTracker(logger *slog.Logger, inactivityTimeout time.Duration) *ServerStateTracker {
	return &ServerStateTracker{
		logger:          logger,
		activeServers:   make(map[string]bool),
		lastActivity:    make(map[string]time.Time),
		inactivityTimer: inactivityTimeout,
	}
}

// SetCallbacks sets the callbacks for server state changes
func (s *ServerStateTracker) SetCallbacks(onActive, onInactive func(serverID string)) {
	s.callbacksMu.Lock()
	defer s.callbacksMu.Unlock()
	s.onServerActive = onActive
	s.onServerInactive = onInactive
}

// UpdateActivity updates the last activity timestamp for a server
func (s *ServerStateTracker) UpdateActivity(serverID string) {
	s.lastActivityMu.Lock()
	defer s.lastActivityMu.Unlock()
	s.lastActivity[serverID] = time.Now()
}

// MarkActive marks a server as active and triggers the callback if it wasn't already active
func (s *ServerStateTracker) MarkActive(serverID string) {
	s.activeServersMu.Lock()
	wasActive := s.activeServers[serverID]
	s.activeServers[serverID] = true
	s.activeServersMu.Unlock()

	// Only trigger callback if server wasn't already active
	if !wasActive {
		s.logger.Debug("Server became active", "serverID", serverID, "reason", "log rotation detected")

		s.callbacksMu.RLock()
		callback := s.onServerActive
		s.callbacksMu.RUnlock()

		if callback != nil {
			go callback(serverID)
		}
	}
}

// MarkInactive marks a server as inactive and triggers the callback if it was active
func (s *ServerStateTracker) MarkInactive(serverID string) {
	s.activeServersMu.Lock()
	wasActive := s.activeServers[serverID]
	s.activeServers[serverID] = false
	s.activeServersMu.Unlock()

	// Only trigger callback if server was actually active
	if wasActive {
		s.logger.Debug("Server became inactive", "serverID", serverID, "reason", "no activity for 10s")

		s.callbacksMu.RLock()
		callback := s.onServerInactive
		s.callbacksMu.RUnlock()

		if callback != nil {
			go callback(serverID)
		}
	}
}

// CheckInactiveServers checks for servers that haven't had activity within the inactivity threshold
// and marks them as inactive. Returns the list of servers that became inactive.
func (s *ServerStateTracker) CheckInactiveServers() []string {
	s.lastActivityMu.RLock()
	now := time.Now()
	inactivityThreshold := s.inactivityTimer

	var inactiveServers []string
	for serverID, lastTime := range s.lastActivity {
		if now.Sub(lastTime) > inactivityThreshold {
			// Check if server is currently marked as active
			s.activeServersMu.RLock()
			isActive := s.activeServers[serverID]
			s.activeServersMu.RUnlock()

			if isActive {
				inactiveServers = append(inactiveServers, serverID)
			}
		}
	}
	s.lastActivityMu.RUnlock()

	// Mark servers as inactive and trigger callbacks
	for _, serverID := range inactiveServers {
		s.MarkInactive(serverID)
	}

	return inactiveServers
}

// IsActive returns whether a server is currently marked as active
func (s *ServerStateTracker) IsActive(serverID string) bool {
	s.activeServersMu.RLock()
	defer s.activeServersMu.RUnlock()
	return s.activeServers[serverID]
}

// GetLastActivity returns the last activity time for a server
func (s *ServerStateTracker) GetLastActivity(serverID string) (time.Time, bool) {
	s.lastActivityMu.RLock()
	defer s.lastActivityMu.RUnlock()
	t, ok := s.lastActivity[serverID]
	return t, ok
}
