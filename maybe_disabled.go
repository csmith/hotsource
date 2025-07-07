//go:build nohotsource

package hotsource

import (
	"net/http"
	"time"
)

// Maybe either creates a hotsource handler, or just returns the fallback
// handler, depending on the value of enable.
func Maybe(fallback http.Handler, dir string, enable bool) http.Handler {
	return fallback
}

// MaybeWithDebounce either creates a hotsource handler with a custom debounce
// duration, or just returns the fallback handler, depending on the value of enable.
func MaybeWithDebounce(fallback http.Handler, dir string, enable bool, debounce time.Duration) http.Handler {
	return fallback
}
