package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	grpcServer "github.com/xxf098/lite-proxy/api/rpc/liteserver"
	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/core"
	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

var (
	port    = flag.Int("p", 8090, "set port")
	test    = flag.String("test", "", "test from command line with subscription link or file")
	conf    = flag.String("config", "", "command line options")
	ping    = flag.Int("ping", 2, "retry times to ping link on startup")
	grpc    = flag.Bool("grpc", false, "start grpc server")
	version = flag.Bool("v", false, "show current version of clash")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("LiteSpeedTest  %s %s %s with %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
		return
	}
	link := ""
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if _, err := utils.CheckLink(arg); err == nil {
			link = arg
			break
		}
	}
	if *test != "" {
		if err := webServer.TestFromCMD(*test, conf); err != nil {
			log.Fatal(err)
		}
		return
	}
	// start grpc server
	if *grpc {
		if err := grpcServer.StartServer(uint16(*port)); err != nil {
			log.Fatalln(err)
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
	c := core.Config{
		LocalHost: "0.0.0.0",
		LocalPort: *port,
		Link:      link,
		Ping:      *ping,
	}
	p, err := core.StartInstance(c)
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
