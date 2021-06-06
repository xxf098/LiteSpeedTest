package main

import (
	"bufio"
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
	wasm = flag.Bool("wasm", false, "start wasm")
)

func main() {
	flag.Parse()
	if isInputFromPipe() {
		r := bufio.NewReader(os.Stdin)
		l, err := r.ReadString(byte('\n'))
		if err == nil {
			link = &l
		}
	}
	if (*test || len(os.Args) < 2) && *link == "" {
		err := webServer.ServeFile(10888)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	if (*wasm || len(os.Args) < 2) && *link == "" {
		err := webServer.ServeWasm(10888)
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

func isInputFromPipe() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fileInfo.Mode()&os.ModeNamedPipe != 0
}
