package main

import (
	"flag"

	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	link = flag.String("link", "", "proxy link")
	port = flag.Int("port", 8090, "local port")
	test = flag.Bool("test", false, "start test")
)

func main() {
	flag.Parse()
	if *test {
		webServer.ServeFile()
		return
	}
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
