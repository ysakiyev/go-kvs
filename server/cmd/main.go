package main

import (
	pb "go-kvs/proto/pb"
	"go-kvs/server/kvs"
	"google.golang.org/grpc"
	"log"
	"net"
)

import g "go-kvs/server/grpc"

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKvsServer(grpcServer, g.NewKvsServer(kvs.New()))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
