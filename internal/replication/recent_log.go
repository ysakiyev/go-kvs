package replication

import (
	"sync"

	go_kvs "go-kvs/api/proto/pb"
)

const DefaultRecentLogSize = 10000

// RecentLog keeps recent commands in memory for catch-up
type RecentLog struct {
	commands []*go_kvs.ReplicationCommand
	capacity int
	startSeq int64 // Sequence number of first command in buffer
	mu       sync.RWMutex
}

func NewRecentLog(capacity int) *RecentLog {
	if capacity <= 0 {
		capacity = DefaultRecentLogSize
	}
	return &RecentLog{
		commands: make([]*go_kvs.ReplicationCommand, 0, capacity),
		capacity: capacity,
		startSeq: 1,
	}
}

// Add appends a command to the recent log
func (r *RecentLog) Add(cmd *go_kvs.ReplicationCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands = append(r.commands, cmd)

	// If buffer is full, remove oldest command
	if len(r.commands) > r.capacity {
		r.commands = r.commands[1:]
		r.startSeq++
	}
}

// GetSince returns commands since the given sequence number
// Returns nil if requested sequence is too old (already evicted)
func (r *RecentLog) GetSince(lastSeq int64) ([]*go_kvs.ReplicationCommand, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If asking for sequence 0, return all commands
	if lastSeq == 0 {
		result := make([]*go_kvs.ReplicationCommand, len(r.commands))
		copy(result, r.commands)
		return result, true
	}

	// Check if requested sequence is too old
	if lastSeq < r.startSeq {
		return nil, false // Too old, can't catch up from memory
	}

	// Calculate offset in buffer
	offset := lastSeq - r.startSeq + 1

	// If offset is beyond current buffer, follower is already caught up
	if offset >= int64(len(r.commands)) {
		return []*go_kvs.ReplicationCommand{}, true
	}

	// Return commands from offset onwards
	result := make([]*go_kvs.ReplicationCommand, len(r.commands)-int(offset))
	copy(result, r.commands[offset:])
	return result, true
}

// GetLatestSequence returns the sequence number of the most recent command
func (r *RecentLog) GetLatestSequence() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.commands) == 0 {
		return 0
	}
	return r.commands[len(r.commands)-1].Sequence
}

// GetStartSequence returns the oldest sequence number still in buffer
func (r *RecentLog) GetStartSequence() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.startSeq
}
