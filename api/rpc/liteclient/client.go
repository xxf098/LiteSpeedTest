package liteclient

import (
	"context"
	"io"
	"log"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"github.com/xxf098/lite-proxy/download"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartClient(addr string, req *pb.TestRequest) ([]*pb.TestReply, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewTestProxyClient(conn)
	ctx := context.Background()
	stream, err := c.StartTest(ctx, req)
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
		log.Println("id: ", reply.Id, reply.Remarks, "ping:", reply.Ping, "avg:", download.ByteCountIECTrim(reply.AvgSpeed), "max:", download.ByteCountIECTrim(reply.MaxSpeed))
		result = append(result, reply)

	}
	return result, nil
}
