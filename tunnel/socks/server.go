package socks

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/tunnel"
)

type Server struct {
	connChan         chan tunnel.Conn
	packetChan       chan tunnel.PacketConn
	underlay         tunnel.Server
	localHost        string
	localPort        int
	timeout          time.Duration
	listenPacketConn net.PacketConn
	mapping          map[string]*PacketConn
	mappingLock      sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, errors.New("socks server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	select {
	case conn := <-s.packetChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("socks server closed")
	}
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	listenPacketConn, err := underlay.AcceptPacket(&Tunnel{})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:         underlay,
		ctx:              ctx,
		cancel:           cancel,
		connChan:         make(chan tunnel.Conn, 32),
		packetChan:       make(chan tunnel.PacketConn, 32),
		localHost:        "127.0.0.1",
		localPort:        8090,
		timeout:          5 * time.Second,
		listenPacketConn: listenPacketConn,
		mapping:          make(map[string]*PacketConn),
	}
	return server, nil
}
