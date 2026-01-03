package replication

import (
	"fmt"
	"sync"

	go_kvs "go-kvs/api/proto/pb"

	"google.golang.org/grpc"
)

type FollowerPool struct {
	connections map[string]go_kvs.ReplicationClient
	mu          sync.RWMutex
}

func NewFollowerPool(followerAddrs []string) (*FollowerPool, error) {
	pool := &FollowerPool{
		connections: make(map[string]go_kvs.ReplicationClient),
	}

	// Connect to each follower
	for _, addr := range followerAddrs {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
		}
		pool.connections[addr] = go_kvs.NewReplicationClient(conn)
	}

	return pool, nil
}
