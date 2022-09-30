package liteclient

import (
	"testing"

	s "github.com/xxf098/lite-proxy/api/rpc/liteserver"
)

func TestStartClient(t *testing.T) {
	go s.StartServer(10999)
	replay, err := StartClient("127.0.0.1:10999")
	if err != nil {
		t.Fatal(err)
	}
	if replay.Message != "ok" {
		t.Fail()
	}

}
