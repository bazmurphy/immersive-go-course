package main

import (
	"flag"
	"log"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/multiple-servers/api"
	"github.com/joho/godotenv"
)

func main() {
	// load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading the .env file")
	}

	// get the database url from the .env file
	databaseURL := os.Getenv("DATABASE_URL")

	port := flag.String("port", "8080", "port to run the api server on")

	flag.Parse()

	api.Run(databaseURL, *port)
}

// go run ./cmd/api-server --port 8081
