package liteserver

import (
	"fmt"
	"net"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"google.golang.org/grpc"
)

type server struct {
	pb.TestProxyServer
}

// stream
func (s *server) StartTest(req *pb.TestRequest, stream pb.TestProxy_StartTestServer) error {
	reply := pb.TestReply{
		Message: req.Name,
	}
	if err := stream.Send(&reply); err != nil {
		return err
	}
	return nil
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
