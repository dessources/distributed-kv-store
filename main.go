package main

import "net/http"

func main() {
	wal, err := NewWal("./leader.wal")
	if err != nil {
		println("Could not initialize write ahead log")
		println(err.Error())
	}
	store, err := NewStore(1024, wal)

	if err != nil {
		println("Could not initialize store")
		println(err.Error())
	}

	server := NewServer(store)
	println("Server Listening...")
	http.ListenAndServe(":8000", server.Handler)

}
