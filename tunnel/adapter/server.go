package adapter

import (
	"bufio"
	"context"
	"errors"
	"net"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/log"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/freedom"
	"github.com/xxf098/lite-proxy/tunnel/http"
	"github.com/xxf098/lite-proxy/tunnel/socks"
	"github.com/xxf098/lite-proxy/utils"
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

// TODO: https socks4
func (s *Server) acceptConnLoop(tcpListener net.Listener) {
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				log.D("exiting")
				return
			default:
				continue
			}
		}

		br := bufio.NewReader(conn)
		b, err := br.Peek(1)
		if err != nil {
			log.Error(common.NewError("failed to detect proxy protocol type").Base(err))
			conn.Close()
			continue
		}
		cc := &common.BufferdConn{Conn: conn, Br: br}
		switch b[0] {
		case 4:
			log.Error(common.NewError("not support proxy protocol type").Base(err))
			conn.Close()
			continue
		case 5:
			s.socksConn <- &freedom.Conn{
				Conn: cc,
			}
		default:
			s.httpConn <- &freedom.Conn{
				Conn: cc,
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
	} else if _, ok := overlay.(*socks.Tunnel); ok {
		s.nextSocks = true
		select {
		case conn := <-s.socksConn:
			return conn, nil
		case <-s.ctx.Done():
			return nil, errors.New("adapter closed")
		}
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
	localHost := ctx.Value("LocalHost").(string)
	localPort := ctx.Value("LocalPort").(int)
	addr := tunnel.NewAddressFromHostPort("tcp", localHost, localPort)
	// tcpListener, err := utils.Listen(ctx, "tcp", addr.String())
	// if err != nil {
	// 	return nil, common.NewError("adapter failed to create tcp listener").Base(err)
	// }
	tcpListeners, err := utils.GetListens(ctx, "tcp", addr.String())
	if err != nil || len(tcpListeners) < 1 {
		return nil, common.NewError("adapter failed to create tcp listener").Base(err)
	}
	udpListener, err := net.ListenPacket("udp", addr.String())
	if err != nil {
		return nil, common.NewError("adapter failed to create tcp listener").Base(err)
	}
	server := &Server{
		tcpListener: tcpListeners[0],
		udpListener: udpListener,
		socksConn:   make(chan tunnel.Conn, 32),
		httpConn:    make(chan tunnel.Conn, 32),
		ctx:         ctx,
		cancel:      cancel,
	}
	for _, v := range tcpListeners {
		go server.acceptConnLoop(v)
	}
	return server, nil
}
