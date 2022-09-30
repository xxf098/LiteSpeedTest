package liteclient

import (
	"context"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartClient(addr string) (*pb.TestReply, error) {
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
	return c.StartTest(ctx, &req)
}
