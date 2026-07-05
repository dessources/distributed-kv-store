package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dessources/distributed-kv-store/pb"
)

// --- LEADER (SERVER) ---

type ReplicationServer struct {
	pb.UnimplementedReplicationServer
	wal *WAL
}

func (s *ReplicationServer) Sync(req *pb.SyncRequest, stream pb.Replication_SyncServer) error {
	cursor := req.StartingOffset
	buf := make([]byte, 4096) // 4KB chunks

	for {
		bytesRead, err := s.wal.f.ReadAt(buf, cursor)
		if bytesRead > 0 {
			chunk := &pb.WalChunk{Data: buf[:bytesRead]}
			if err := stream.Send(chunk); err != nil {
				return err
			}
			cursor += int64(bytesRead)
		} else if err == io.EOF {
			// Tailing the file. If no new bytes are present, sleep briefly.
			time.Sleep(10 * time.Millisecond)
		} else if err != nil {
			return err
		}
	}
}

func StartGRPCServer(port string, wal *WAL) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterReplicationServer(grpcServer, &ReplicationServer{wal: wal})

	fmt.Println("gRPC Replication Leader listening on", port)
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}

// --- FOLLOWER (CLIENT) ---

func StartReplicationClient(leaderAddr string, wal *WAL, store *Store) {
	// 1. Determine local WAL size to know where to resume (Crash Tolerance)
	info, err := wal.f.Stat()
	if err != nil {
		panic(err)
	}
	startingOffset := info.Size()

	// 2. Connect to Leader
	// (Using insecure for local network simplicity)
	conn, err := grpc.Dial(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewReplicationClient(conn)
	req := &pb.SyncRequest{StartingOffset: startingOffset}
	
	stream, err := client.Sync(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to Leader %s. Syncing from offset %d...\n", leaderAddr, startingOffset)

	// 3. Pipe the incoming gRPC binary chunks into a standard io.Reader
	pr, pw := io.Pipe()

	// Routine A: Feed the gRPC chunks into the pipe and write them to our local Follower WAL
	go func() {
		for {
			chunk, err := stream.Recv()
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			
			// Persist to local disk so we don't lose data on reboot
			wal.f.Write(chunk.Data)
			wal.f.Sync() 

			// Feed into the in-memory parser
			pw.Write(chunk.Data)
		}
	}()

	// Routine B: Run our existing binary parser against the pipe!
	err = RecoverFromReader(pr, store)
	if err != nil {
		fmt.Println("Replication stream disconnected:", err)
	}
}
