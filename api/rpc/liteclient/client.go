package liteclient

import (
	"context"
	"io"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartClient(addr string) ([]*pb.TestReply, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewTestProxyClient(conn)
	ctx := context.Background()
	req := pb.TestRequest{
		Name: "ok",
	}
	stream, err := c.StartTest(ctx, &req)
	if err != nil {
		return nil, err
	}
	result := []*pb.TestReply{}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			stream.CloseSend()
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, reply)

	}
	return result, nil
}
