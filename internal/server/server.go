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
	kvs       *kvs.Kvs
	streamMgr *replication.StreamManager
	isLeader  bool
	go_kvs.UnimplementedGoKvsServer
}

func NewKvsServer(kvs *kvs.Kvs, streamMgr *replication.StreamManager, isLeader bool) *KvsServer {
	return &KvsServer{
		kvs:       kvs,
		streamMgr: streamMgr,
		isLeader:  isLeader,
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

	// Broadcast to followers via streams
	if k.streamMgr != nil {
		cmd := command.New("set", request.Key, request.Val)
		cmdBytes, _ := cmd.Serialize()
		k.streamMgr.Broadcast(cmdBytes)
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

	// Broadcast to followers via streams
	if k.streamMgr != nil {
		cmd := command.New("del", request.Key, "")
		cmdBytes, _ := cmd.Serialize()
		k.streamMgr.Broadcast(cmdBytes)
	}

	return &go_kvs.EmptyResponse{}, nil
}

func (k *KvsServer) Keys(ctx context.Context, request *go_kvs.EmptyRequest) (*go_kvs.KeysResponse, error) {
	keys := k.kvs.Keys()
	return &go_kvs.KeysResponse{Keys: keys}, nil
}
