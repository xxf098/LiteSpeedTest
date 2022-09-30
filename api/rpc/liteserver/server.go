package liteserver

import (
	"context"
	"fmt"
	"net"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"google.golang.org/grpc"
)

type server struct {
	pb.TestProxyServer
}

func (s *server) StartTest(_ context.Context, req *pb.TestRequest) (*pb.TestReply, error) {
	replay := pb.TestReply{
		Message: req.Name,
	}
	return &replay, nil
}

func StartServer(port uint16) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterTestProxyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
