package server

import (
	go_kvs "go-kvs/api/proto/pb"
	"go-kvs/internal/replication"

	"github.com/rs/zerolog/log"
)

type LeaderStreamServer struct {
	streamMgr *replication.StreamManager
	go_kvs.UnimplementedReplicationServer
}

func NewLeaderStreamServer(streamMgr *replication.StreamManager) *LeaderStreamServer {
	return &LeaderStreamServer{
		streamMgr: streamMgr,
	}
}

// StreamReplication handles follower connections and streams commands to them
func (s *LeaderStreamServer) StreamReplication(req *go_kvs.FollowerInfo, stream go_kvs.Replication_StreamReplicationServer) error {
	followerID := req.FollowerId
	log.Info().Msgf("Follower %s connected from %s", followerID, req.FollowerAddr)

	// Create channel for this follower (buffered to handle bursts)
	cmdChan := make(chan *go_kvs.ReplicationCommand, 100)

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
