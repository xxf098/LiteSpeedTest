package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/common"
	N "github.com/xxf098/lite-proxy/common/net"
	"github.com/xxf098/lite-proxy/common/pool"
	"github.com/xxf098/lite-proxy/log"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/utils"
)

// proxy http/scocks to vmess
const Name = "PROXY"

type Proxy struct {
	sources []tunnel.Server
	sink    tunnel.Client
	ctx     context.Context
	cancel  context.CancelFunc
	pool    *utils.WorkerPool
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
	if p.pool != nil {
		p.pool.Stop()
	}
	return nil
}

// forward from socks/http connection to vmess/trojan
// TODO: bypass cn
func (p *Proxy) relayConnLoop() {
	pool := utils.WorkerPool{
		WorkerFunc: func(inbound tunnel.Conn) error {
			defer inbound.Close()
			addr := inbound.Metadata().Address
			var outbound net.Conn
			var err error
			start := time.Now()
			if addr.IP != nil && N.IsPrivateAddress(addr.IP) {
				networkType := addr.NetworkType
				if networkType == "" {
					networkType = "tcp"
				}
				add := net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
				outbound, err = net.Dial(networkType, add)
			} else {
				outbound, err = p.sink.DialConn(addr, nil)
			}
			if err != nil {
				log.Error(common.NewError("proxy failed to dial connection").Base(err))
				return err
			}
			elapsed := fmt.Sprintf("%dms", time.Since(start).Milliseconds())
			log.D("connect to:", addr, elapsed)
			defer outbound.Close()
			// relay
			return relay(outbound, inbound)
		},
		MaxWorkersCount:       2000,
		LogAllErrors:          false,
		MaxIdleWorkerDuration: 2 * time.Minute,
	}
	p.pool = &pool
	pool.Start()
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
				pool.Serve(inbound)
			}
		}(source)
	}
}

func relay(leftConn, rightConn net.Conn) error {
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
	err = <-ch
	return err
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
