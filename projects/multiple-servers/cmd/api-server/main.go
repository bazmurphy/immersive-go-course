package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/api"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("api-server environment: DATABASE_URL not provided")
		os.Exit(1)
	}

	port := flag.String("port", "8080", "port to run the api server on")

	flag.Parse()

	if *port == "" {
		fmt.Println("api server flag: port not provided")
		os.Exit(1)
	}

	api.Run(databaseURL, *port)
}
