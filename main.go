package main

import (
	"flag"
	"log"
	"os"

	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	link = flag.String("link", "", "add subscription link")
	port = flag.Int("port", 8090, "set port")
	test = flag.Bool("test", false, "start batch test")
)

func main() {
	flag.Parse()
	if (*test || len(os.Args) < 2) && *link == "" {
		err := webServer.ServeFile(10888)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	if *link == "" {
		if len(os.Args) > 1 {
			for i := 1; i < len(os.Args); i++ {
				arg := os.Args[i]
				if _, err := utils.CheckLink(arg); err == nil {
					link = &arg
					break
				}
			}

		}
		if *link == "" {
			err := webServer.ServeFile(*port)
			if err != nil {
				log.Fatalln(err)
			}
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
