package main

import (
	"flag"
	"fmt"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/api"
)

func main() {
	path := flag.String("path", "", "path to the static files")
	port := flag.String("port", "8080", "port to run the static files server on")

	flag.Parse()

	fmt.Println("path", *path, "port", *port)

	api.Run(*path, *port)
}

// go run ./cmd/static-server --path assets --port 8082
