//go:build !nohotsource

package hotsource

import (
	"log/slog"
	"net/http"
	"time"
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

// MaybeWithDebounce either creates a hotsource handler with a custom debounce
// duration, or just returns the fallback handler, depending on the value of enable.
func MaybeWithDebounce(fallback http.Handler, dir string, enable bool, debounce time.Duration) http.Handler {
	if !enable {
		return fallback
	}

	h, err := NewWithDebounce(dir, debounce)
	if err != nil {
		slog.Debug("Unable to create hotsource handler", "error", err, "source", "hostsource")
		return fallback
	}

	return h
}
