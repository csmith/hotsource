//go:build !nohotsource

package hotsource

import (
	"log/slog"
	"net/http"
)

// Maybe either creates a hotsource handler, or just returns the fallback
// handler, depending on the value of enable.
func Maybe(fallback http.Handler, dir string, enable bool) http.Handler {
	if !enable {
		return fallback
	}

	h, err := New(dir)
	if err != nil {
		slog.Debug("Unable to create hotsource handler", "error", err, "source", "hostsource")
		return fallback
	}

	return h
}
