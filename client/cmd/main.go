package main

import (
	"context"
	"fmt"
	pb "go-kvs/proto/pb"
	"google.golang.org/grpc"
	"log"
)
import g "go-kvs/client/grpc"

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := g.NewKvsClient(conn)

	_, _ = client.Set(context.Background(), &pb.KeyValRequest{Key: "33", Val: "123"})
	res, _ := client.Get(context.Background(), &pb.KeyRequest{Key: "33"})
	fmt.Println(res.Value)

}
