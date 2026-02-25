package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	pb "github.com/rajesh/kvstore/proto"
	"github.com/rajesh/kvstore/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getConnection(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func dialPeers(peersStr string) map[uint32]*server.Peer {
	peers := make(map[uint32]*server.Peer)
	if peersStr == "" {
		return peers
	}
	for _, p := range strings.Split(peersStr, ",") {
		parts := strings.SplitN(p, "@", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid peer format %q: expected id@host:port", p)
		}
		peerID, err := strconv.Atoi(parts[0])
		if err != nil || peerID == 0 {
			log.Fatalf("invalid peer id %q: must be a non-zero integer", parts[0])
		}
		conn, err := getConnection(parts[1])
		if err != nil {
			log.Fatalf("failed to connect to peer %s: %v", parts[1], err)
		}
		id := uint32(peerID)
		peers[id] = &server.Peer{
			ID:     id,
			Client: pb.NewRaftServiceClient(conn),
			Conn:   conn,
		}
	}
	return peers
}

func main() {
	// Usage:
	// go run main.go -id=1 -port=50051 -peers=2@localhost:50052,3@localhost:50053

	id := flag.Int("id", 0, "Unique ID for this server (required, non-zero)")
	port := flag.Int("port", 0, "Port to listen on (required, non-zero)")
	peersStr := flag.String("peers", "", "Comma-separated list of peer servers in format id@host:port")
	flag.Parse()

	if *id == 0 {
		log.Fatal("-id is required and must be non-zero")
	}
	if *port == 0 {
		log.Fatal("-port is required and must be non-zero")
	}

	// Dial peers before constructing the server so the election timer starts
	// with a complete peer list — no two-phase initialization.
	peers := dialPeers(*peersStr)
	for _, p := range peers {
		defer p.Conn.Close()
	}

	raftServer := server.NewRaftServer(uint32(*id), peers)
	grpcServer := grpc.NewServer()
	raftServer.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}
	log.Printf("Node %d listening on port %d", *id, *port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
