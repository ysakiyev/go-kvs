package grpc

import (
	context "context"
	pb "go-kvs/proto/pb"
	"go-kvs/server/kvs"
)

type KvsServer struct {
	kvs *kvs.Kvs
	pb.UnimplementedGoKvsServer
}

func NewKvsServer(kvs *kvs.Kvs) *KvsServer {
	return &KvsServer{
		kvs: kvs,
	}
}

func (k *KvsServer) Get(ctx context.Context, request *pb.KeyRequest) (*pb.ValResponse, error) {
	res, err := k.kvs.Get(request.Key)
	if err != nil {
		return nil, err
	}
	return &pb.ValResponse{
		Value: res,
	}, nil
}

func (k *KvsServer) Set(ctx context.Context, request *pb.KeyValRequest) (*pb.EmptyResponse, error) {
	err := k.kvs.Set(request.Key, request.Val)
	if err != nil {
		return nil, err
	}
	return &pb.EmptyResponse{}, nil
}

func (k *KvsServer) Del(ctx context.Context, request *pb.KeyRequest) (*pb.EmptyResponse, error) {
	err := k.kvs.Del(request.Key)
	if err != nil {
		return nil, err
	}
	return &pb.EmptyResponse{}, nil
}
