package stats

import (
	"io"
)

type SizeStatWriter struct {
	Counter *Counter
	Writer  io.Writer
}

func (s SizeStatWriter) Write(p []byte) (int, error) {
	n, err := s.Writer.Write(p)
	s.Counter.Add(int64(n))
	return n, err
}

type SizeStatReader struct {
	Counter *Counter
	Reader  io.Reader
}

func (s SizeStatReader) Read(p []byte) (int, error) {
	n, err := s.Reader.Read(p)
	s.Counter.Add(int64(n))
	return n, err
}
