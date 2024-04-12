package api

import "fmt"

func Run(databaseURL string, port string) {
	fmt.Println("Hello from /api/api.go")
	fmt.Println("databaseURL", databaseURL, "port", port)
}
