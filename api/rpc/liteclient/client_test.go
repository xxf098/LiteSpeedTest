package liteclient

import (
	"testing"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	s "github.com/xxf098/lite-proxy/api/rpc/liteserver"
)

func TestStartClient(t *testing.T) {
	go s.StartServer(10999)
	req := pb.TestRequest{
		GroupName: "ok",
	}
	reply, err := StartClient("127.0.0.1:10999", &req)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply) < 1 {
		t.Fail()
	}
	if reply[0].GroupName != "ok" {
		t.Fail()
	}

}
