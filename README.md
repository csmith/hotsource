# Hot source: live HTML reloading for Go

Hot source provides an easy way to make your web frontend automatically reload
during development when you make changes to it.

## Why?

Automatic reloading makes development a lot easier. Being able to have a
browser window visible and update as you make changes in an IDE saves a lot
of time switching back and forth, manually reloading, and so on. Also, it's just
fun.

This sort of live-reloading comes as standard in a lot of JavaScript frameworks,
but in Go we're mostly stuck with tools that rerun the entire binary, and
ignore anything going on in the frontend.

## What?

Hot source provides a `http.Handler` that injects a cookie and a bit of JavaScript
into any request for a HTML file. The JavaScript opens a websocket connection,
and will cause the browser to reload if it's told to via websocket.

The handler also tracks any resources requested with its cookie set (i.e., from
the same browser session). It adds watchers to those files, and if they're
updated it will send the refresh command to the relevant websockets.

## … Maybe?

But you don't always want live reloading, so there's also a `Maybe` func that
can enable or disable hot source based on a boolean. That makes it easy to hook
it up to a flag, and only enable live reloading in dev mode.

If you use the `Maybe` func you can even set the `nohotsource` build tag and
none of hot source's dependencies will end up in your binary.

## How?

Create your `http.Handler` as you usually would, and then wrap it in a call to
`hotsource.Maybe`:

```go
package main

import (
	"embed"
	"flag"
	"github.com/csmith/hotsource"
	"http"
	"io/fs"
)

//go:embed frontend/*.html frontend/*.js frontend/*.png
var frontendFS embed.FS

var hotReload = flag.Bool("hot-reload", false, "Enable hot reloading")

func main() {
	flag.Parse()
	
	normalHandler := http.FileServerFS(fs.Sub(frontendFS, "frontend"))
	hotHandler := hotsource.Maybe(normalHandler, "frontend", *hotReload)

	http.ListenAndServ(":8080", hotHandler)
}
```

## But…?

There are some caveats:

- Watchers are never removed, and are always for single files. This will leak
  resources over time, but it's intended for development use not long term
  deployment.
- The way resources are tracked might not be perfect. In particular, if one
  HTML page loads another HTML page (either via an iframe or something like
  HTMX), it will end up with a new session ID. I think this will still work
  properly, but it hasn't been tested.

## Help?

This is quite rough and ready. Any contributions to tidy it up, add more
features, improve the documentation, etc, are more than welcome.

If you use hot source and encounter a bug or problem feel free to raise an
issue.