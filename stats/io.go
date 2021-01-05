package stats

import (
	"net"
)

// type SizeStatWriter struct {
// 	Counter *Counter
// 	Writer  io.Writer
// }

// func (s SizeStatWriter) Write(p []byte) (int, error) {
// 	n, err := s.Writer.Write(p)
// 	s.Counter.Add(int64(n))
// 	return n, err
// }

// type SizeStatReader struct {
// 	Counter *Counter
// 	Reader  io.Reader
// }

// func (s SizeStatReader) Read(p []byte) (int, error) {
// 	n, err := s.Reader.Read(p)
// 	s.Counter.Add(int64(n))
// 	return n, err
// }

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
