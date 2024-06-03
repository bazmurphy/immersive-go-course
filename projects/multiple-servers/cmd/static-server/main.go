package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/static"
)

func main() {
	path := flag.String("path", "", "path to the static files")
	port := flag.String("port", "8080", "port to run the static files server on")

	flag.Parse()

	if *path == "" {
		fmt.Println("static server flag: path not provided")
		os.Exit(1)
	}

	if *port == "" {
		fmt.Println("static server flag: port not provided")
		os.Exit(1)
	}

	static.Run(*path, *port)
}
