package follower

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	go_kvs "go-kvs/api/proto/pb"
	"go-kvs/pkg/kvs"
	"go-kvs/pkg/kvs/command"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type StreamClient struct {
	nodeID       string
	leaderAddr   string
	kvs          *kvs.Kvs
	lastSequence int64
	seqFile      string
}

func NewStreamClient(nodeID, leaderAddr string, kvs *kvs.Kvs) *StreamClient {
	seqFile := fmt.Sprintf(".%s.seq", nodeID)

	client := &StreamClient{
		nodeID:       nodeID,
		leaderAddr:   leaderAddr,
		kvs:          kvs,
		lastSequence: 0,
		seqFile:      seqFile,
	}

	// Load last sequence from file
	client.loadLastSequence()

	return client
}

// loadLastSequence loads the last applied sequence from disk
func (f *StreamClient) loadLastSequence() {
	data, err := os.ReadFile(f.seqFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error().Err(err).Msg("Failed to read sequence file")
		}
		f.lastSequence = 0
		return
	}

	seq, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse sequence file")
		f.lastSequence = 0
		return
	}

	f.lastSequence = seq
	log.Info().Msgf("Loaded last sequence: %d", f.lastSequence)
}

// saveLastSequence saves the last applied sequence to disk
func (f *StreamClient) saveLastSequence() error {
	data := fmt.Sprintf("%d\n", f.lastSequence)
	return os.WriteFile(f.seqFile, []byte(data), 0644)
}

// ConnectToLeader connects to leader and receives command stream
func (f *StreamClient) ConnectToLeader() {
	for {
		log.Info().Msgf("Connecting to leader at %s (last_seq=%d)...", f.leaderAddr, f.lastSequence)

		conn, err := grpc.Dial(f.leaderAddr, grpc.WithInsecure())
		if err != nil {
			log.Error().Err(err).Msg("Failed to dial leader, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}

		client := go_kvs.NewReplicationClient(conn)
		stream, err := client.StreamReplication(context.Background(), &go_kvs.FollowerInfo{
			FollowerId:   f.nodeID,
			FollowerAddr: "",
			LastSequence: f.lastSequence, // Send last sequence for catch-up
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to start stream, retrying in 2s...")
			conn.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Info().Msg("Connected to leader, receiving stream")

		// Receive commands from stream (including catch-up commands)
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
				// Update last sequence
				f.lastSequence = cmd.Sequence

				// Persist sequence every 10 commands (or could be every command)
				if f.lastSequence%10 == 0 {
					if err := f.saveLastSequence(); err != nil {
						log.Error().Err(err).Msg("Failed to save sequence")
					}
				}

				log.Debug().Msgf("Applied command seq=%d", cmd.Sequence)
			}
		}

		// Save sequence on disconnect
		if err := f.saveLastSequence(); err != nil {
			log.Error().Err(err).Msg("Failed to save sequence on disconnect")
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
