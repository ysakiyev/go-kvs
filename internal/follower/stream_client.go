package follower

import (
	"context"
	"time"

	go_kvs "go-kvs/api/proto/pb"
	"go-kvs/pkg/kvs"
	"go-kvs/pkg/kvs/command"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type StreamClient struct {
	nodeID     string
	leaderAddr string
	kvs        *kvs.Kvs
}

func NewStreamClient(nodeID, leaderAddr string, kvs *kvs.Kvs) *StreamClient {
	return &StreamClient{
		nodeID:     nodeID,
		leaderAddr: leaderAddr,
		kvs:        kvs,
	}
}

// ConnectToLeader connects to leader and receives command stream
func (f *StreamClient) ConnectToLeader() {
	for {
		log.Info().Msgf("Connecting to leader at %s...", f.leaderAddr)

		conn, err := grpc.Dial(f.leaderAddr, grpc.WithInsecure())
		if err != nil {
			log.Error().Err(err).Msg("Failed to dial leader, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}

		client := go_kvs.NewReplicationClient(conn)
		stream, err := client.StreamReplication(context.Background(), &go_kvs.FollowerInfo{
			FollowerId:   f.nodeID,
			FollowerAddr: "", // Could add own address here if needed
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to start stream, retrying in 2s...")
			conn.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Info().Msg("Connected to leader, receiving stream")

		// Receive commands from stream
		for {
			cmd, err := stream.Recv()
			if err != nil {
				log.Error().Err(err).Msg("Stream error, reconnecting in 2s...")
				conn.Close()
				time.Sleep(2 * time.Second)
				break // Break inner loop to reconnect
			}

			// Apply command to local KVS
			if err := f.applyCommand(cmd); err != nil {
				log.Error().Err(err).Msgf("Failed to apply command seq=%d", cmd.Sequence)
			} else {
				log.Debug().Msgf("Applied command seq=%d", cmd.Sequence)
			}
		}
	}
}

// applyCommand deserializes and applies a command to the local KVS
func (f *StreamClient) applyCommand(cmd *go_kvs.ReplicationCommand) error {
	// Deserialize command
	c, err := command.Deserialize(cmd.Command)
	if err != nil {
		return err
	}

	// Apply to local KVS
	switch c.Cmd {
	case "set":
		return f.kvs.Set(c.Key, c.Val)
	case "del":
		return f.kvs.Del(c.Key)
	default:
		log.Warn().Msgf("Unknown command type: %s", c.Cmd)
		return nil
	}
}
