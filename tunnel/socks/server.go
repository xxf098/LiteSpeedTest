package socks

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/tunnel"
)

type Server struct {
	connChan         chan net.Conn
	packetChan       chan net.PacketConn
	underlay         tunnel.Server
	localHost        string
	localPort        int
	timeout          time.Duration
	listenPacketConn net.PacketConn
	// mapping          map[string]*PacketConn
	mappingLock sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *Server) AcceptConn(tunnel.Tunnel) (net.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, errors.New("socks server closed")
	}
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*tunnel.Server, error) {
	return nil, nil
}
