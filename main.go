package main

import (
	"net/http"
)

func main() {
	wal, err := NewWal("/home/pianodessources/Projects/distributed-kv-store/log.wal")
	if err != nil {
		println("could not initialize write-ahead log:")
		println(err)
	}
	store := NewStore(100, wal)
	wal.Recover(store)
	server := NewServer(store)

	http.ListenAndServe(":8000", server.Handler)
}
