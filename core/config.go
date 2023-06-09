package core

type Config struct {
	LocalHost string
	LocalPort int
	Link      string
	Ping      int
	PingCh    chan<- int64 // ping result chan
}
