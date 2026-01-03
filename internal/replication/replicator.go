package replication

import (
	"context"
	"fmt"
	"sync"
	"time"

	go_kvs "go-kvs/api/proto/pb"

	"github.com/rs/zerolog/log"
)

type Replicator struct {
	pool *FollowerPool
}

func NewReplicator(pool *FollowerPool) *Replicator {
	return &Replicator{
		pool: pool,
	}
}

func (r *Replicator) Replicate(cmd []byte) error {
	// Send to all followers in parallel
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(r.pool.connections))

	for addr, client := range r.pool.connections {
		wg.Add(1)
		go func(addr string, client go_kvs.ReplicationClient) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := client.ReplicateCommand(ctx, &go_kvs.ReplicateRequest{
				Command: cmd,
			})

			if err != nil || !resp.Success {
				errorsChan <- fmt.Errorf("replication to %s failed: %v", addr, err)
			}
		}(addr, client)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		log.Error().Err(err).Msg("Replication error")
		// For now, just log errors (don't block client)
	}

	return nil
}
