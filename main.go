package main

import (
	"flag"
	"log"
	"os"

	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	port = flag.Int("p", 8090, "set port")
)

func main() {
	flag.Parse()
	link := ""
	for _, arg := range os.Args {
		if _, err := utils.CheckLink(arg); err == nil {
			link = arg
			break
		}
	}
	if link == "" {
		if len(os.Args) < 2 {
			*port = 10888
		}
		if err := webServer.ServeFile(*port); err != nil {
			log.Fatalln(err)
		}
		return
	}
	c := Config{
		LocalHost: "127.0.0.1",
		LocalPort: *port,
		Link:      link,
	}
	if p, err := startInstance(c); err != nil {
		log.Fatalln(err)
	} else {
		p.Run()
	}

}
