package server

import (
	"context"
	"go-kvs/api/proto/pb"
	"go-kvs/internal/replication"
	"go-kvs/pkg/kvs"
	"go-kvs/pkg/kvs/command"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type KvsServer struct {
	kvs        *kvs.Kvs
	replicator *replication.Replicator // NEW
	isLeader   bool                    // NEW
	go_kvs.UnimplementedGoKvsServer
}

func NewKvsServer(kvs *kvs.Kvs, replicator *replication.Replicator, isLeader bool) *KvsServer {
	return &KvsServer{
		kvs:        kvs,
		replicator: replicator,
		isLeader:   isLeader,
	}
}

func (k *KvsServer) Get(ctx context.Context, request *go_kvs.KeyRequest) (*go_kvs.ValResponse, error) {
	res, err := k.kvs.Get(request.Key)
	if err != nil {
		return nil, err
	}

	return &go_kvs.ValResponse{
		Value: res,
	}, nil
}

func (k *KvsServer) Set(ctx context.Context, request *go_kvs.KeyValRequest) (*go_kvs.EmptyResponse, error) {
	// Check if this node is the leader
	if !k.isLeader {
		return nil, status.Error(codes.FailedPrecondition, "not leader")
	}

	// Apply to local KVS first
	err := k.kvs.Set(request.Key, request.Val)
	if err != nil {
		return nil, err
	}

	// Replicate to followers (async, fire-and-forget for now)
	if k.replicator != nil {
		cmd := command.New("set", request.Key, request.Val)
		cmdBytes, _ := cmd.Serialize()
		go k.replicator.Replicate(cmdBytes)
	}

	return &go_kvs.EmptyResponse{}, nil
}

func (k *KvsServer) Del(ctx context.Context, request *go_kvs.KeyRequest) (*go_kvs.EmptyResponse, error) {
	// Check if this node is the leader
	if !k.isLeader {
		return nil, status.Error(codes.FailedPrecondition, "not leader")
	}

	// Apply to local KVS first
	err := k.kvs.Del(request.Key)
	if err != nil {
		return nil, err
	}

	// Replicate to followers (async, fire-and-forget for now)
	if k.replicator != nil {
		cmd := command.New("del", request.Key, "")
		cmdBytes, _ := cmd.Serialize()
		go k.replicator.Replicate(cmdBytes)
	}

	return &go_kvs.EmptyResponse{}, nil
}
