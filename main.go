package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":6543", GalleryHandler{})
	if err != nil {
		log.Fatal(err)
	}
}

type GalleryHandler struct{}

func (h GalleryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "", "/":
		index(w, r)
	case "/image":
		serveFile(w, r)
	default:
		serveGallery(w, r)
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome")
}
