package vmess

import (
	"context"
	"crypto/tls"
	"net"

	C "github.com/xxf098/lite-proxy/constant"
)

type TLSConfig struct {
	Host           string
	SkipCertVerify bool
	SessionCache   tls.ClientSessionCache
	NextProtos     []string
}

func StreamTLSConn(conn net.Conn, cfg *TLSConfig) (net.Conn, error) {
	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.SkipCertVerify,
		ClientSessionCache: cfg.SessionCache,
		NextProtos:         cfg.NextProtos,
	}

	tlsConn := tls.Client(conn, tlsConfig)
	ctx, cancel := context.WithTimeout(context.Background(), C.DefaultTLSTimeout)
	defer cancel()
	err := tlsConn.HandshakeContext(ctx)
	return tlsConn, err
}
