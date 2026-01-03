package main

import (
	"flag"
	"fmt"
	"net"

	pb "go-kvs/api/proto/pb"
	"go-kvs/internal/config"
	"go-kvs/internal/follower"
	"go-kvs/internal/replication"
	g "go-kvs/internal/server"
	"go-kvs/internal/server/middleware"
	"go-kvs/pkg/kvs"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Parse command-line flags
	cfg := parseFlags()

	// Initialize KVS with node-specific WAL file
	walFile := fmt.Sprintf("wal-%s.log", cfg.NodeID)
	kvsInstance, err := kvs.New(walFile)
	if err != nil {
		log.Fatal().Msgf("Failed to init KVS: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal().Msgf("Failed to listen on %s: %v", cfg.Address, err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryServerLoggingInterceptor))

	if cfg.IsLeader {
		// Leader setup
		log.Info().Msgf("Starting as LEADER on %s", cfg.Address)

		// Create stream manager for followers
		streamMgr := replication.NewStreamManager()

		// Register client-facing KVS service
		kvsServer := g.NewKvsServer(kvsInstance, streamMgr, true)
		pb.RegisterGoKvsServer(grpcServer, kvsServer)

		// Register replication service for follower connections
		leaderStreamServer := g.NewLeaderStreamServer(streamMgr)
		pb.RegisterReplicationServer(grpcServer, leaderStreamServer)
	} else {
		// Follower setup
		log.Info().Msgf("Starting as FOLLOWER on %s", cfg.Address)
		log.Info().Msgf("Leader: %s", cfg.LeaderAddr)

		// Register client-facing KVS service (read-only, will reject writes)
		kvsServer := g.NewKvsServer(kvsInstance, nil, false)
		pb.RegisterGoKvsServer(grpcServer, kvsServer)

		// Start stream client to connect to leader (runs in background)
		streamClient := follower.NewStreamClient(cfg.NodeID, cfg.LeaderAddr, kvsInstance)
		go streamClient.ConnectToLeader()
	}

	log.Info().Msgf("Server listening on %s", cfg.Address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Msgf("Failed to serve: %v", err)
	}
}

func parseFlags() *config.ServerConfig {
	nodeID := flag.String("node-id", "node1", "Unique node identifier")
	isLeader := flag.Bool("leader", false, "Run as leader")
	port := flag.String("port", "50051", "Port to listen on")
	leaderAddr := flag.String("leader-addr", "", "Leader address (follower only)")

	flag.Parse()

	cfg := &config.ServerConfig{
		NodeID:   *nodeID,
		IsLeader: *isLeader,
		Address:  fmt.Sprintf("localhost:%s", *port),
	}

	if !*isLeader {
		if *leaderAddr == "" {
			log.Fatal().Msg("Follower must specify --leader-addr")
		}
		cfg.LeaderAddr = *leaderAddr
	}

	return cfg
}
