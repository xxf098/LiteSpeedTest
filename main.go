package main

import (
	"flag"
)

var (
	link = flag.String("link", "", "proxy link")
)

func main() {
	flag.Parse()
	if *link == "" {
		return
	}
	c := Config{
		LocalHost: "127.0.0.1",
		LocalPort: 8090,
		Link:      *link,
	}
	p, err := startInstance(c)
	if err != nil {
		return
	}
	p.Run()
}
