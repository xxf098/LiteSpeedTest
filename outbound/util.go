package outbound

import (
	"bytes"
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/component/socks5"
	C "github.com/xxf098/lite-proxy/constant"
)

const (
	tcpTimeout = 5 * time.Second
)

var (
	globalClientSessionCache tls.ClientSessionCache
	once                     sync.Once
)

func getClientSessionCache() tls.ClientSessionCache {
	once.Do(func() {
		globalClientSessionCache = tls.NewLRUClientSessionCache(128)
	})
	return globalClientSessionCache
}

func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}

func serializesSocksAddr(metadata *C.Metadata) []byte {
	var buf [][]byte
	aType := uint8(metadata.AddrType)
	p, _ := strconv.Atoi(metadata.DstPort)
	port := []byte{uint8(p >> 8), uint8(p & 0xff)}
	switch metadata.AddrType {
	case socks5.AtypDomainName:
		len := uint8(len(metadata.Host))
		host := []byte(metadata.Host)
		buf = [][]byte{{aType, len}, host, port}
	case socks5.AtypIPv4:
		host := metadata.DstIP.To4()
		buf = [][]byte{{aType}, host, port}
	case socks5.AtypIPv6:
		host := metadata.DstIP.To16()
		buf = [][]byte{{aType}, host, port}
	}
	return bytes.Join(buf, nil)
}
