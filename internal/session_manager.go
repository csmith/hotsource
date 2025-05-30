package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/fsnotify/fsnotify"
)

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	watcher  *fsnotify.Watcher
}

func NewSessionManager() (*SessionManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	sm := &SessionManager{
		sessions: make(map[string]*Session),
		watcher:  watcher,
	}

	go sm.watchFiles()
	go sm.startReaper()

	return sm, nil
}

func (sm *SessionManager) GetOrCreateSession(sessionID string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.mu.Lock()
		session.lastSeen = time.Now()
		session.mu.Unlock()
		return session
	}

	session := &Session{
		id:              sessionID,
		connections:     make(map[string]*websocket.Conn),
		lastSeen:        time.Now(),
		watchedPaths:    make(map[string]bool),
		hasModification: false,
	}
	sm.sessions[sessionID] = session

	slog.Debug("Created session", "session_id", sessionID, "source", "hostsource")
	return session
}

func (sm *SessionManager) AddWebSocketToSession(sessionID string, connID string, conn *websocket.Conn) {
	session := sm.GetOrCreateSession(sessionID)

	session.AddConnection(connID, conn)
	slog.Debug("Added websocket to session", "session_id", sessionID, "conn_id", connID, "source", "hostsource")
}

func (sm *SessionManager) RemoveWebSocketFromSession(sessionID string, connID string) {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	session.RemoveConnection(connID)
	slog.Debug("Removed websocket from session", "session_id", sessionID, "conn_id", connID, "source", "hostsource")
}

func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.Close()
		delete(sm.sessions, sessionID)
		slog.Debug("Removed session", "session_id", sessionID, "source", "hostsource")
	}
}

func (sm *SessionManager) startReaper() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.reapSessions()
	}
}

func (sm *SessionManager) reapSessions() {
	var toRemove []string

	sm.mu.RLock()
	for sessionID, session := range sm.sessions {
		if session.Stale() {
			toRemove = append(toRemove, sessionID)
		}
	}
	sm.mu.RUnlock()

	for _, sessionID := range toRemove {
		slog.Debug("Reaping inactive session", "session_id", sessionID, "source", "hostsource")
		sm.RemoveSession(sessionID)
	}
}

func (sm *SessionManager) watchFiles() {
	for {
		select {
		case event, ok := <-sm.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				slog.Debug("File changed, notifying relevant sessions", "path", event.Name, "operation", event.Op.String(), "source", "hostsource")
				sm.notifySessionsForPath(event.Name)
			}
		case err, ok := <-sm.watcher.Errors:
			if !ok {
				return
			}
			slog.Debug("File watcher error", "error", err, "source", "hostsource")
		}
	}
}

func (sm *SessionManager) notifySessionsForPath(path string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.sessions {
		if session.Watching(path) {
			session.Notify()
		}
	}
}

func (sm *SessionManager) SessionExists(sessionID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.sessions[sessionID]
	return exists
}

func (sm *SessionManager) AddWatchPath(sessionID, path string) error {
	session := sm.GetOrCreateSession(sessionID)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			path = indexPath
		} else {
			return fmt.Errorf("directory has no index.html")
		}
	}

	session.Watch(path)

	if err := sm.watcher.Add(path); err != nil {
		return err
	}
	slog.Debug("Adding path to watcher", "path", path, "source", "hostsource")

	return nil
}
