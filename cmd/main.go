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

	raftServer := server.NewRaftServer(uint32(*id))
	grpcServer := grpc.NewServer()
	raftServer.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}
	log.Printf("Node %d listening on port %d", *id, *port)

	var peers map[string]string
	peers = make(map[string]string)

	if *peersStr != "" {
		parts := strings.Split(*peersStr, ",")
		for _, p := range parts {
			sub := strings.Split(p, "@") // ["2", "localhost:50052"]
			id := (sub[0])
			peers[id] = sub[1] // "2"
		}
	}

	fmt.Println("ID:", *id)
	fmt.Println("Port:", *port)
	fmt.Println("Peers:", peers)

	peersMap := make(map[uint32]*server.Peer)

	for id, peerAddr := range peers {
		conn, err := getConnection(peerAddr)
		if err != nil {
			fmt.Printf("Failed to connect to peer %s: %v\n", peerAddr, err)
			continue
		}
		defer conn.Close()

		client := pb.NewRaftServiceClient(conn)
		id, _ := strconv.Atoi(id)
		peersMap[uint32(id)] = &server.Peer{
			ID:     uint32(id),
			Client: client,
			Conn:   conn,
		}
	}

	raftServer.SetPeers(peersMap)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
