package main

import (
	"flag"
)

var (
	link = flag.String("link", "", "proxy link")
	port = flag.Int("port", 8090, "local port")
)

func main() {
	flag.Parse()
	if *link == "" {
		return
	}
	c := Config{
		LocalHost: "127.0.0.1",
		LocalPort: *port,
		Link:      *link,
	}
	p, err := startInstance(c)
	if err != nil {
		return
	}
	p.Run()
}
