package server

import (
	gokvs "go-kvs/api/proto/pb"
	"go-kvs/internal/replication"

	"github.com/rs/zerolog/log"
)

type LeaderStreamServer struct {
	streamMgr *replication.StreamManager
	gokvs.UnimplementedReplicationServer
}

func NewLeaderStreamServer(streamMgr *replication.StreamManager) *LeaderStreamServer {
	return &LeaderStreamServer{
		streamMgr: streamMgr,
	}
}

// StreamReplication handles follower connections and streams commands to them
func (s *LeaderStreamServer) StreamReplication(req *gokvs.FollowerInfo, stream gokvs.Replication_StreamReplicationServer) error {
	followerID := req.FollowerId
	lastSeq := req.LastSequence
	log.Info().Msgf("Follower %s connected (last_seq=%d)", followerID, lastSeq)

	// Step 1: Catch-up - replay missed commands
	missedCommands, canCatchUp := s.streamMgr.GetMissedCommands(lastSeq)

	if !canCatchUp {
		log.Warn().Msgf("Follower %s too far behind (last_seq=%d), needs manual sync", followerID, lastSeq)
		// In production, you'd trigger snapshot transfer here
		// For now, we'll warn but still continue with live stream
	} else if len(missedCommands) > 0 {
		log.Info().Msgf("Replaying %d missed commands to follower %s", len(missedCommands), followerID)

		// Send missed commands
		for _, cmd := range missedCommands {
			if err := stream.Send(cmd); err != nil {
				log.Error().Err(err).Msgf("Failed to send catch-up command to %s", followerID)
				return err
			}
			log.Debug().Msgf("Catch-up: sent seq=%d to follower %s", cmd.Sequence, followerID)
		}

		log.Info().Msgf("Follower %s caught up successfully", followerID)
	} else {
		log.Info().Msgf("Follower %s is already up-to-date", followerID)
	}

	// Step 2: Start live streaming
	// Create channel for this follower (buffered to handle bursts)
	cmdChan := make(chan *gokvs.ReplicationCommand, 100)

	// Register follower
	s.streamMgr.Register(followerID, cmdChan)
	defer s.streamMgr.Unregister(followerID)

	// Send commands from channel to stream
	for cmd := range cmdChan {
		if err := stream.Send(cmd); err != nil {
			log.Error().Err(err).Msgf("Failed to send to follower %s, disconnecting", followerID)
			return err
		}
		log.Debug().Msgf("Sent seq=%d to follower %s", cmd.Sequence, followerID)
	}

	return nil
}
