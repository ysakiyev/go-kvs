package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pb "go-kvs/proto/pb"
	g "go-kvs/server/grpc"
	"go-kvs/server/kvs"
	"google.golang.org/grpc"
	"net"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Msgf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGoKvsServer(grpcServer, g.NewKvsServer(kvs.New()))

	log.Info().Msg("Listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Msgf("Failed to serve: %v", err)
	}
}
