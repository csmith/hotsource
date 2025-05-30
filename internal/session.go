package internal

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Session struct {
	id              string
	connections     map[string]*websocket.Conn
	lastSeen        time.Time
	watchedPaths    map[string]bool
	hasModification bool
	mu              sync.RWMutex
}

func (s *Session) Notify() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.connections) > 0 {
		for connID, conn := range s.connections {
			if err := conn.Write(context.Background(), websocket.MessageText, []byte("reload")); err != nil {
				slog.Debug("Failed to send message to connection", "session_id", s.id, "conn_id", connID, "error", err, "source", "hostsource")
				delete(s.connections, connID)
			}
		}
		s.hasModification = false
	} else {
		// No active connections, mark for later
		s.hasModification = true
	}
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for connID, conn := range s.connections {
		_ = conn.Close(websocket.StatusNormalClosure, "session closed")
		delete(s.connections, connID)
	}
}

func (s *Session) Watch(path string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.watchedPaths[path] = true
}

func (s *Session) Watching(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.watchedPaths[path]
}

func (s *Session) AddConnection(connID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connections[connID] = conn
	if s.hasModification {
		go func() {
			if err := conn.Write(context.Background(), websocket.MessageText, []byte("reload")); err != nil {
				slog.Debug("Failed to send immediate reload to new connection", "session_id", s.id, "conn_id", connID, "error", err, "source", "hostsource")
			}
			s.mu.Lock()
			s.hasModification = false
			s.mu.Unlock()
		}()
	}
}

func (s *Session) RemoveConnection(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.connections, connID)
}

func (s *Session) Stale() bool {
	return len(s.connections) == 0 && time.Since(s.lastSeen) > 5*time.Minute
}
