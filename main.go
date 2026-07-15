package main

import (
	"log"
	"net/http"
)

func NewFileServerHandler(dir string) http.Handler {
	return http.FileServer(http.Dir(dir))
}

func main() {
	const port = ":2000"
	handler := NewFileServerHandler("./static")

	http.Handle("/", handler)

	log.Printf("listening on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
