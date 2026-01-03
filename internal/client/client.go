package client

import (
	context "context"
	"go-kvs/api/proto/pb"
	"google.golang.org/grpc"
)

type KvsClient struct {
	client go_kvs.GoKvsClient
}

func NewKvsClient(conn *grpc.ClientConn) *KvsClient {
	return &KvsClient{client: go_kvs.NewGoKvsClient(conn)}
}

func (k *KvsClient) Get(ctx context.Context, in *go_kvs.KeyRequest, opts ...grpc.CallOption) (*go_kvs.ValResponse, error) {
	return k.client.Get(ctx, in, opts...)
}

func (k *KvsClient) Set(ctx context.Context, in *go_kvs.KeyValRequest, opts ...grpc.CallOption) (*go_kvs.EmptyResponse, error) {
	return k.client.Set(ctx, in, opts...)
}

func (k *KvsClient) Del(ctx context.Context, in *go_kvs.KeyRequest, opts ...grpc.CallOption) (*go_kvs.EmptyResponse, error) {
	return k.client.Del(ctx, in, opts...)
}

func (k *KvsClient) Keys(ctx context.Context, in *go_kvs.EmptyRequest, opts ...grpc.CallOption) (*go_kvs.KeysResponse, error) {
	return k.client.Keys(ctx, in, opts...)
}
