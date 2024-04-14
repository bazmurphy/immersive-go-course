package static

import (
	"fmt"
	"net/http"
)

func Run(path string, port string) {
	// fmt.Println("hello from /static/static.go")
	// fmt.Println("path", path, "port", port)

	fileServer := http.FileServer(http.Dir(path))

	http.Handle("/", fileServer)

	fmt.Printf("static server running on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}
