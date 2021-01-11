package stats

import (
	"net"
	"runtime"
)

type StatsConn struct {
	net.Conn
	CounterDown *Counter
	CounterUp   *Counter
}

func (s StatsConn) Read(b []byte) (n int, err error) {
	n, err = s.Conn.Read(b)
	s.CounterDown.Add(int64(n))
	return
}

func (s StatsConn) Write(b []byte) (n int, err error) {
	n, err = s.Conn.Write(b)
	s.CounterUp.Add(int64(n))
	return
}

func NewStatsConn(c net.Conn) StatsConn {
	return StatsConn{
		Conn:        c,
		CounterDown: DefaultManager.GetCounter(DownProxy),
		CounterUp:   DefaultManager.GetCounter(UpProxy),
	}
}

func NewConn(c net.Conn) net.Conn {
	if runtime.GOOS != "android" {
		return c
	}
	return NewStatsConn(c)
}

type StatsPacketConn struct {
	net.PacketConn
	CounterDown *Counter
	CounterUp   *Counter
}

func (pc StatsPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = pc.PacketConn.ReadFrom(p)
	pc.CounterDown.Add(int64(n))
	return
}

func (pc StatsPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = pc.PacketConn.WriteTo(p, addr)
	pc.CounterUp.Add(int64(n))
	return
}

func NewStatsPacketConn(pc net.PacketConn) StatsPacketConn {
	return StatsPacketConn{
		PacketConn:  pc,
		CounterDown: DefaultManager.GetCounter(DownProxy),
		CounterUp:   DefaultManager.GetCounter(UpProxy),
	}
}
