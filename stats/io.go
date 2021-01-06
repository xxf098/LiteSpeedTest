package stats

import (
	"net"
)

type StatsConn struct {
	net.Conn
	CounterDown *Counter
	CounterUp   *Counter
}

func (s StatsConn) Read(b []byte) (n int, err error) {
	n, err = s.Conn.Read(b)
	s.CounterDown.Add(int64(n))
	return n, err
}

func (s StatsConn) Write(b []byte) (n int, err error) {
	n, err = s.Conn.Write(b)
	s.CounterUp.Add(int64(n))
	return n, err
}

func NewStatsConn(c net.Conn) StatsConn {
	return StatsConn{
		Conn:        c,
		CounterDown: DefaultManager.GetCounter(DownProxy),
		CounterUp:   DefaultManager.GetCounter(UpProxy),
	}
}
