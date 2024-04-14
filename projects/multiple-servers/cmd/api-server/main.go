package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/api"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading the .env file: %v", err)
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")

	port := flag.String("port", "8080", "port to run the api server on")

	flag.Parse()

	api.Run(databaseURL, *port)
}
