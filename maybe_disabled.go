//go:build nohotsource

package hotsource

import (
	"net/http"
)

// Maybe either creates a hotsource handler, or just returns the fallback
// handler, depending on the value of enable.
func Maybe(fallback http.Handler, dir string, enable bool) http.Handler {
	return fallback
}
