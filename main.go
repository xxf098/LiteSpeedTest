package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	port = flag.Int("p", 8090, "set port")
	test = flag.String("test", "", "test from command line with subscription link or file")
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
	if *test != "" {
		if err := webServer.TestFromCMD(*test); err != nil {
			log.Fatal(err)
		}
		return
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
	p, err := startInstance(c)
	if err != nil {
		log.Fatalln(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigs)
	go func() {
		<-sigs
		p.Close()
	}()
	p.Run()
}
