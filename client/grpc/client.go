package grpc

import (
	context "context"
	pb "go-kvs/proto/pb"
	"google.golang.org/grpc"
)

type KvsClient struct {
	client pb.GoKvsClient
}

func NewKvsClient(conn *grpc.ClientConn) *KvsClient {
	return &KvsClient{client: pb.NewGoKvsClient(conn)}
}

func (k *KvsClient) Get(ctx context.Context, in *pb.KeyRequest, opts ...grpc.CallOption) (*pb.ValResponse, error) {
	return k.client.Get(ctx, in, opts...)
}

func (k *KvsClient) Set(ctx context.Context, in *pb.KeyValRequest, opts ...grpc.CallOption) (*pb.EmptyResponse, error) {
	return k.client.Set(ctx, in, opts...)
}

func (k *KvsClient) Del(ctx context.Context, in *pb.KeyRequest, opts ...grpc.CallOption) (*pb.EmptyResponse, error) {
	return k.client.Del(ctx, in, opts...)
}
