package liteserver

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	"github.com/xxf098/lite-proxy/web"
	"google.golang.org/grpc"
)

type server struct {
	pb.TestProxyServer
}

// stream
func (s *server) StartTest(req *pb.TestRequest, stream pb.TestProxy_StartTestServer) error {
	// check data
	links, err := web.ParseLinks(req.Subscription)
	if err != nil {
		return err
	}
	groupName := req.GroupName
	if len(groupName) < 1 {
		groupName = "Default"
	}
	concurrency := req.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	timeout := time.Duration(req.Timeout)
	if timeout < 15 {
		timeout = 15
	}
	speedTestMode := "all"
	if req.SpeedTestMode == pb.SpeedTestMode_pingonly {
		speedTestMode = "pingonly"
	} else if req.SpeedTestMode == pb.SpeedTestMode_speedonly {
		speedTestMode = "speedonly"
	}
	sortMethod := "none"
	if req.SortMethod == pb.SortMethod_ping {
		sortMethod = "ping"
	} else if req.SortMethod == pb.SortMethod_rping {
		sortMethod = "rping"
	} else if req.SortMethod == pb.SortMethod_rping {
		sortMethod = "rping"
	}
	// config
	p := web.ProfileTest{
		Writer:      nil,
		MessageType: web.ALLTEST,
		Links:       links,
		Options: &web.ProfileTestOptions{
			GroupName:     groupName,
			SpeedTestMode: speedTestMode,
			PingMethod:    "googleping",
			SortMethod:    sortMethod,
			Concurrency:   int(concurrency),
			TestMode:      2,
			Timeout:       timeout * time.Second,
			Language:      "en",
			FontSize:      24,
		},
	}

	nodeChan, err := p.TestAll(stream.Context(), nil)
	count := 0
	linkCount := len(links)
	for count < linkCount {
		node := <-nodeChan
		reply := pb.TestReply{
			Id:        int32(node.Id),
			GroupName: node.Group,
			Remarks:   node.Remarks,
			Protocol:  node.Protocol,
			Ping:      node.Ping,
			AvgSpeed:  node.AvgSpeed,
			MaxSpeed:  node.MaxSpeed,
			IsOk:      node.IsOk,
			Traffic:   node.Traffic,
			Link:      node.Link,
		}
		if err := stream.Send(&reply); err != nil {
			return err
		}
		count += 1
	}
	return nil
}

func StartServer(port uint16) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	log.Printf("start grpc server at %s", lis.Addr().String())
	s := grpc.NewServer()
	pb.RegisterTestProxyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
