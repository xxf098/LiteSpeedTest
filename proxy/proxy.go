package proxy

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/xxf098/lite-proxy/common"
	N "github.com/xxf098/lite-proxy/common/net"
	"github.com/xxf098/lite-proxy/common/pool"
	"github.com/xxf098/lite-proxy/log"
	"github.com/xxf098/lite-proxy/tunnel"
)

// proxy http/scocks to vmess
const Name = "PROXY"

type Proxy struct {
	sources []tunnel.Server
	sink    tunnel.Client
	ctx     context.Context
	cancel  context.CancelFunc
}

func (p *Proxy) Run() error {
	p.relayConnLoop()
	// p.relayPacketLoop()
	<-p.ctx.Done()
	return nil
}

func (p *Proxy) Close() error {
	p.cancel()
	p.sink.Close()
	for _, source := range p.sources {
		source.Close()
	}
	return nil
}

// forward from socks/http connection to vmess/trojan
func (p *Proxy) relayConnLoop() {
	for _, source := range p.sources {
		go func(source tunnel.Server) {
			for {
				inbound, err := source.AcceptConn(nil)
				if err != nil {
					select {
					case <-p.ctx.Done():
						log.D("exiting")
						return
					default:
					}
					log.Error(common.NewError("failed to accept connection").Base(err))
					continue
				}
				go func(inbound tunnel.Conn) {
					defer inbound.Close()
					outbound, err := p.sink.DialConn(inbound.Metadata().Address, nil)
					if err != nil {
						log.Error(common.NewError("proxy failed to dial connection").Base(err))
						return
					}
					log.D("connect to:", inbound.Metadata().Address)
					defer outbound.Close()
					// relay
					// relay(inbound, outbound)
					errChan := make(chan error, 2)
					copyConn := func(a, b net.Conn) {
						buf := pool.Get(pool.RelayBufferSize)
						_, err := io.CopyBuffer(a, b, buf)
						pool.Put(buf)
						a.SetReadDeadline(time.Now())
						errChan <- err
						return
					}
					go copyConn(inbound, outbound)
					go copyConn(outbound, inbound)
					select {
					case err = <-errChan:
						if err != nil {
							log.E(err.Error())
							return
						}
					case <-p.ctx.Done():
						log.D("shutting down conn relay")
						return
					}
				}(inbound)
			}
		}(source)
	}
}

func relay(leftConn, rightConn net.Conn) {
	ch := make(chan error)

	go func() {
		buf := pool.Get(pool.RelayBufferSize)
		// Wrapping to avoid using *net.TCPConn.(ReadFrom)
		// See also https://github.com/Dreamacro/clash/pull/1209
		_, err := io.CopyBuffer(N.WriteOnlyWriter{Writer: leftConn}, N.ReadOnlyReader{Reader: rightConn}, buf)
		if err != nil {
			log.E(err.Error())
		}
		pool.Put(buf)
		leftConn.SetReadDeadline(time.Now())
		ch <- err
	}()

	buf := pool.Get(pool.RelayBufferSize)
	_, err := io.CopyBuffer(N.WriteOnlyWriter{Writer: rightConn}, N.ReadOnlyReader{Reader: leftConn}, buf)
	if err != nil {
		log.E(err.Error())
	}
	pool.Put(buf)
	rightConn.SetReadDeadline(time.Now())
	<-ch
}

func NewProxy(ctx context.Context, cancel context.CancelFunc, sources []tunnel.Server, sink tunnel.Client) *Proxy {
	return &Proxy{
		sources: sources,
		sink:    sink,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// A Dialer is a means to establish a connection.
// Custom dialers should also implement ContextDialer.
type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
}
