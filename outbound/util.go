package outbound

import (
	"crypto/tls"
	"net"
	"sync"
	"time"
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
