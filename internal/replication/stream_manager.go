package replication

import (
	"sync"

	go_kvs "go-kvs/api/proto/pb"

	"github.com/rs/zerolog/log"
)

type StreamManager struct {
	streams   map[string]chan *go_kvs.ReplicationCommand
	recentLog *RecentLog
	mu        sync.RWMutex
	sequence  int64
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams:   make(map[string]chan *go_kvs.ReplicationCommand),
		recentLog: NewRecentLog(DefaultRecentLogSize),
		sequence:  0,
	}
}

// Register a new follower stream
func (sm *StreamManager) Register(followerID string, ch chan *go_kvs.ReplicationCommand) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.streams[followerID] = ch
	log.Info().Msgf("Follower %s registered, total followers: %d", followerID, len(sm.streams))
}

// Unregister a follower stream
func (sm *StreamManager) Unregister(followerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if ch, exists := sm.streams[followerID]; exists {
		close(ch)
		delete(sm.streams, followerID)
		log.Info().Msgf("Follower %s unregistered, total followers: %d", followerID, len(sm.streams))
	}
}

// Broadcast command to all connected followers
func (sm *StreamManager) Broadcast(cmdBytes []byte) {
	sm.mu.Lock()
	sm.sequence++
	seq := sm.sequence
	sm.mu.Unlock()

	cmd := &go_kvs.ReplicationCommand{
		Command:  cmdBytes,
		Sequence: seq,
	}

	// Add to recent log for catch-up
	sm.recentLog.Add(cmd)

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for followerID, ch := range sm.streams {
		select {
		case ch <- cmd:
			// Successfully sent
			log.Debug().Msgf("Sent seq=%d to follower %s", seq, followerID)
		default:
			// Channel full, log warning but don't block
			log.Warn().Msgf("Follower %s channel full (seq=%d), command may be dropped", followerID, seq)
		}
	}
}

// GetMissedCommands returns commands since lastSeq for catch-up
// Returns (commands, canCatchUp). If canCatchUp is false, too many commands missed.
func (sm *StreamManager) GetMissedCommands(lastSeq int64) ([]*go_kvs.ReplicationCommand, bool) {
	return sm.recentLog.GetSince(lastSeq)
}

// GetFollowerCount returns number of connected followers
func (sm *StreamManager) GetFollowerCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.streams)
}
