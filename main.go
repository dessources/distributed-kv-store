package main

import (
	"flag"
	"net/http"
)

func main() {
	role := flag.String("role", "leader", "Node role: leader or follower")
	port := flag.String("port", ":8000", "HTTP port for this node")
	grpcPort := flag.String("grpc-port", ":9000", "gRPC port for replication")
	leaderAddr := flag.String("leader-addr", "localhost:9000", "gRPC address of the leader (for followers)")
	walPath := flag.String("wal", "log.wal", "Path to WAL file")
	flag.Parse()

	wal, err := NewWal(*walPath)
	if err != nil {
		println("could not initialize write-ahead log:")
		panic(err)
	}

	store := NewStore(100, wal)
	
	// Recover existing local state
	if err := wal.Recover(store); err != nil {
		println("recovery error:", err.Error())
	}

	if *role == "leader" {
		// Boot the gRPC Server for followers to connect to
		go StartGRPCServer(*grpcPort, wal)
	} else if *role == "follower" {
		// Connect to the Leader and sync the WAL
		go StartReplicationClient(*leaderAddr, wal, store)
	} else {
		panic("invalid role. use 'leader' or 'follower'")
	}

	// Both leaders and followers run the HTTP API
	server := NewServer(store)
	server.Addr = *port
	
	println("HTTP Server listening on", *port)
	if err := http.ListenAndServe(*port, server.Handler); err != nil {
		panic(err)
	}
}
