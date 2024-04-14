package main

import (
	"flag"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/static"
)

func main() {
	path := flag.String("path", "", "path to the static files")
	port := flag.String("port", "8080", "port to run the static files server on")

	flag.Parse()

	// fmt.Println("path", *path, "port", *port)

	static.Run(*path, *port)
}

// go run ./cmd/static-server --path assets --port 8082
