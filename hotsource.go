package hotsource

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/csmith/hotsource/internal"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// New creates a new hotsource handler that will serve files from the given
// directory. HTML files will have javascript injected that will cause them
// to reload when any requested resource is changed on disk.
func New(dir string) (http.Handler, error) {
	sm, err := internal.NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("unable to create session manager: %w", err)
	}

	h := &handler{
		handler: http.FileServerFS(os.DirFS(dir)),
		sm:      sm,
		dir:     dir,
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("GET /_reload", handleWebSocket(sm))
	return mux, nil
}

type handler struct {
	handler http.Handler
	sm      *internal.SessionManager
	dir     string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := internal.NewResponseInterceptor()

	h.handler.ServeHTTP(rec, r)

	contentType := rec.Header().Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		sessionID := uuid.New().String()
		http.SetCookie(w, &http.Cookie{
			Name:  "live-reload-session",
			Value: sessionID,
			Path:  "/",
		})

		h.sm.GetOrCreateSession(sessionID)

		requestedPath := filepath.Join(h.dir, r.URL.Path)
		if err := h.sm.AddWatchPath(sessionID, requestedPath); err != nil {
			slog.Debug("Could not watch requested path", "path", requestedPath, "error", err, "source", "hostsource")
		}

		rec.Write([]byte(fmt.Sprintf(`<script>(()=>{const sessionId='%s';function connect(){const ws=new WebSocket('ws://'+location.host+'/_reload');ws.onopen=()=>ws.send(sessionId);ws.onmessage=()=>location.reload();ws.onclose=()=>setTimeout(connect,1000);ws.onerror=()=>setTimeout(connect,1000);}connect();})();</script>`, sessionID)))
		rec.Header().Set("Content-Length", fmt.Sprintf("%d", rec.Body.Len()))
	} else if cookie, err := r.Cookie("live-reload-session"); err == nil {
		requestedPath := filepath.Join(h.dir, r.URL.Path)
		h.sm.GetOrCreateSession(cookie.Value)
		if err := h.sm.AddWatchPath(cookie.Value, requestedPath); err != nil {
			slog.Debug("Could not watch requested path", "path", requestedPath, "error", err, "source", "hostsource")
		}
	}

	rec.WriteTo(w)
}

func handleWebSocket(sessions *internal.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			slog.Debug("Failed to accept websocket connection", "error", err, "source", "hostsource")
			return
		}

		slog.Debug("Live reload websocket connection opened", "remote_addr", r.RemoteAddr, "source", "hostsource")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		_, msg, err := conn.Read(ctx)
		if err != nil {
			slog.Debug("Failed to read session ID", "error", err, "source", "hostsource")
			conn.Close(websocket.StatusInternalError, "")
			return
		}

		sessionID := string(msg)
		if !sessions.SessionExists(sessionID) {
			slog.Debug("Unknown session ID, telling client to reload", "session_id", sessionID, "source", "hostsource")
			conn.Write(ctx, websocket.MessageText, []byte("reload"))
			conn.Close(websocket.StatusNormalClosure, "reload")
			return
		}

		connID := uuid.New().String()
		sessions.AddWebSocketToSession(sessionID, connID, conn)
		defer sessions.RemoveWebSocketFromSession(sessionID, connID)

		<-ctx.Done()
		conn.Close(websocket.StatusNormalClosure, "")
	}
}
