package liteclient

import (
	"testing"

	pb "github.com/xxf098/lite-proxy/api/rpc/lite"
	s "github.com/xxf098/lite-proxy/api/rpc/liteserver"
)

func TestStartClient(t *testing.T) {
	go s.StartServer(10999)
	req := pb.TestRequest{
		GroupName:     "Default",
		SpeedTestMode: pb.SpeedTestMode_all,
		PingMethod:    pb.PingMethod_googleping,
		SortMethod:    pb.SortMethod_rspeed,
		Concurrency:   2,
		TestMode:      2,
		Subscription:  "https://raw.githubusercontent.com/freefq/free/master/v2",
		Language:      "en",
		FontSize:      24,
		Theme:         "rainbow",
		Timeout:       10,
		OutputMode:    0,
	}
	reply, err := StartClient("127.0.0.1:10999", &req)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply) < 1 {
		t.Fail()
	}
}
