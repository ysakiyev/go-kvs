package server

import (
	"context"

	go_kvs "go-kvs/api/proto/pb"
	"go-kvs/pkg/kvs"
	"go-kvs/pkg/kvs/command"
)

type FollowerServer struct {
	kvs *kvs.Kvs
	go_kvs.UnimplementedReplicationServer
}

func NewFollowerServer(kvs *kvs.Kvs) *FollowerServer {
	return &FollowerServer{
		kvs: kvs,
	}
}

func (f *FollowerServer) ReplicateCommand(ctx context.Context, req *go_kvs.ReplicateRequest) (*go_kvs.ReplicateResponse, error) {
	// Deserialize command
	cmd, err := command.Deserialize(req.Command)
	if err != nil {
		return &go_kvs.ReplicateResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Apply to local KVS
	switch cmd.Cmd {
	case "set":
		err = f.kvs.Set(cmd.Key, cmd.Val)
	case "del":
		err = f.kvs.Del(cmd.Key)
	default:
		return &go_kvs.ReplicateResponse{
			Success: false,
			Error:   "unknown command",
		}, nil
	}

	if err != nil {
		return &go_kvs.ReplicateResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &go_kvs.ReplicateResponse{Success: true}, nil
}
