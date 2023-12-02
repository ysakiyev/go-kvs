package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	pb "go-kvs/api/proto/pb"
	g "go-kvs/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"os"
	"strings"
)

func main() {
	conn, dialErr := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if dialErr != nil {
		log.Fatal().Msgf("Failed to dial: %v", dialErr)
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
				if st, ok := status.FromError(err); ok {
					fmt.Printf("Error: %s\n", st.Message())
				}
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
			_, err := client.Set(context.Background(), &pb.KeyValRequest{Key: key, Val: val})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					fmt.Printf("Error: %s\n", st.Message())
				}
				continue
			}

		case "del":
			if len(parts) != 2 {
				fmt.Println("Invalid 'del' command. Usage: del {key}")
				continue
			}
			key := parts[1]
			_, err := client.Del(context.Background(), &pb.KeyRequest{Key: key})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					fmt.Printf("Error: %s\n", st.Message())
				}
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
