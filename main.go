package main

import (
	"flag"
	"log"
	"os"

	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	link = flag.String("link", "", "proxy link")
	port = flag.Int("port", 8090, "local port")
	test = flag.Bool("test", false, "start test")
)

func main() {
	flag.Parse()
	if *test || len(os.Args) < 2 {
		err := webServer.ServeFile()
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	if *link == "" {
		if len(os.Args) > 1 {
			arg := os.Args[1]
			if _, err := utils.CheckLink(os.Args[1]); err == nil {
				link = &arg
			}
		}
		if *link == "" {
			return
		}
	}
	c := Config{
		LocalHost: "127.0.0.1",
		LocalPort: *port,
		Link:      *link,
	}
	p, err := startInstance(c)
	if err != nil {
		log.Fatalln(err)
	}
	p.Run()
}
