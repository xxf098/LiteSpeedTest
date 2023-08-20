package outbound

import (
	"fmt"
	"testing"

	C "github.com/xxf098/lite-proxy/constant"
)

func TestSerializesSocksAddr(t *testing.T) {
	metadata := C.Metadata{
		NetWork: 3,
		Type:    7,
		DstPort: "443",
		Host:    "tj.trojanfree.com",
	}
	buf := serializesSocksAddr(&metadata)
	fmt.Println(buf)
}
