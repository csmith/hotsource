package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/csmith/hotsource"
)

//go:embed data/*
var dataFS embed.FS

var hotReload = flag.Bool("hot-reload", false, "Enable hot reloading")
var port = flag.String("port", "8080", "Port to listen on")

func main() {
	flag.Parse()

	sub, err := fs.Sub(dataFS, "data")
	if err != nil {
		log.Fatal(err)
	}

	normalHandler := http.FileServerFS(sub)
	hotHandler := hotsource.Maybe(normalHandler, "data", *hotReload)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting server on http://localhost%s\n", addr)
	if *hotReload {
		fmt.Println("Hot reloading is enabled! Try editing the HTML, CSS, or JS files in the data folder.")
	} else {
		fmt.Println("Hot reloading is disabled. Use -hot-reload flag to enable it.")
	}

	log.Fatal(http.ListenAndServe(addr, hotHandler))
}