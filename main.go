package main

import (
	"net/http"
)

func main() {
	store := NewStore(100)
	server := NewServer(store)

	http.ListenAndServe(":8000", server.Handler)
}
