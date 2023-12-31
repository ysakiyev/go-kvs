package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pb "go-kvs/api/proto/pb"
	g "go-kvs/internal/server"
	"go-kvs/internal/server/middleware"
	"go-kvs/pkg/kvs"
	"google.golang.org/grpc"
	"net"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Msgf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryServerLoggingInterceptor))
	kvsS, err := kvs.New("wal.log")
	if err != nil {
		log.Fatal().Msgf("Failed to init KVS: %v", err)
	}
	pb.RegisterGoKvsServer(grpcServer, g.NewKvsServer(kvsS))

	log.Info().Msg("Listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Msgf("Failed to serve: %v", err)
	}
}
