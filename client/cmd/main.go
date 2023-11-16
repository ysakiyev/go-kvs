package main

import (
	"bufio"
	"context"
	"fmt"
	g "go-kvs/client/grpc"
	pb "go-kvs/proto/pb"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := g.NewKvsClient(conn)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()

		// Trim leading and trailing whitespaces
		command = strings.TrimSpace(command)

		// Parse the command
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue // Empty command, prompt again
		}

		// Check the command and validate arguments
		switch parts[0] {
		case "get":
			if len(parts) != 2 {
				fmt.Println("Invalid 'get' command. Usage: get {key}")
				continue
			}
			key := parts[1]
			res, err := client.Get(context.Background(), &pb.KeyRequest{Key: key})
			if err != nil {
				continue
			}
			fmt.Println(res.Value)

		case "set":
			if len(parts) != 3 {
				fmt.Println("Invalid 'set' command. Usage: set {key} {val}")
				continue
			}
			key := parts[1]
			val := parts[2]
			_, err = client.Set(context.Background(), &pb.KeyValRequest{Key: key, Val: val})
			if err != nil {
				continue
			}

		case "del":
			if len(parts) != 2 {
				fmt.Println("Invalid 'del' command. Usage: del {key}")
				continue
			}
			key := parts[1]
			_, err = client.Del(context.Background(), &pb.KeyRequest{Key: key})
			if err != nil {
				continue
			}

		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
			return

		default:
			fmt.Println("Invalid command. Valid commands are: get, set, del, exit")
		}
	}
}
