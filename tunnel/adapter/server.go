package adapter

import (
	"context"
	"errors"
	"net"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/log"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/freedom"
	"github.com/xxf098/lite-proxy/tunnel/http"
)

type Server struct {
	tcpListener net.Listener
	udpListener net.PacketConn
	socksConn   chan tunnel.Conn
	httpConn    chan tunnel.Conn
	nextSocks   bool
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *Server) acceptConnLoop() {
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				log.D("exiting")
				return
			default:
				continue
			}
		}
		rewindConn := common.NewRewindConn(conn)
		rewindConn.SetBufferSize(16)
		buf := [3]byte{}
		_, err = rewindConn.Read(buf[:])
		rewindConn.Rewind()
		rewindConn.StopBuffering()
		if err != nil {
			log.Error(common.NewError("failed to detect proxy protocol type").Base(err))
			continue
		}
		if buf[0] == 5 && s.nextSocks {
			log.D("socks5 connection")
			s.socksConn <- &freedom.Conn{
				Conn: rewindConn,
			}
		} else {
			log.D("http connection")
			s.httpConn <- &freedom.Conn{
				Conn: rewindConn,
			}
		}
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	if _, ok := overlay.(*http.Tunnel); ok {
		select {
		case conn := <-s.httpConn:
			return conn, nil
		case <-s.ctx.Done():
			return nil, errors.New("adapter closed")
		}
		// } else if _, ok := overlay.(*socks.Tunnel); ok {
		// 	s.nextSocks = true
		// 	select {
		// 	case conn := <-s.socksConn:
		// 		return conn, nil
		// 	case <-s.ctx.Done():
		// 		return nil, errors.New("adapter closed")
		// 	}
	} else {
		panic("invalid overlay")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return &freedom.PacketConn{
		UDPConn: s.udpListener.(*net.UDPConn),
	}, nil
}

func (s *Server) Close() error {
	s.cancel()
	s.tcpListener.Close()
	return s.udpListener.Close()
}

func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", 8090)
	tcpListener, err := net.Listen("tcp", addr.String())
	if err != nil {
		return nil, common.NewError("adapter failed to create tcp listener").Base(err)
	}
	udpListener, err := net.ListenPacket("udp", addr.String())
	if err != nil {
		return nil, common.NewError("adapter failed to create tcp listener").Base(err)
	}
	server := &Server{
		tcpListener: tcpListener,
		udpListener: udpListener,
		socksConn:   make(chan tunnel.Conn, 32),
		httpConn:    make(chan tunnel.Conn, 32),
		ctx:         ctx,
		cancel:      cancel,
	}
	go server.acceptConnLoop()
	return server, nil
}
