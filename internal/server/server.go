package server

import (
	context "context"
	"go-kvs/api/proto/pb"
	"go-kvs/pkg/kvs"
)

type KvsServer struct {
	kvs *kvs.Kvs
	go_kvs.UnimplementedGoKvsServer
}

func NewKvsServer(kvs *kvs.Kvs) *KvsServer {
	return &KvsServer{
		kvs: kvs,
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
	err := k.kvs.Set(request.Key, request.Val)
	if err != nil {
		return nil, err
	}
	return &go_kvs.EmptyResponse{}, nil
}

func (k *KvsServer) Del(ctx context.Context, request *go_kvs.KeyRequest) (*go_kvs.EmptyResponse, error) {
	err := k.kvs.Del(request.Key)
	if err != nil {
		return nil, err
	}
	return &go_kvs.EmptyResponse{}, nil
}
