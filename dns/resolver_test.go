package dns

import (
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {
	c := Config{
		Main: []NameServer{
			{
				Net:  "udp",
				Addr: "8.8.8.8:53",
			},
			{
				Net:  "udp",
				Addr: "223.5.5.5:53",
			},
		},
	}
	r := NewResolver(c)
	// TODO: resolve ipv4
	ip, err := r.ResolveIP("www.github.com")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("ip: %s\n", ip.String())
}
