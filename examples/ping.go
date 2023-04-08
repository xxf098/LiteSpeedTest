package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xxf098/lite-proxy/download"
	"github.com/xxf098/lite-proxy/web"
)

func main() {
	async := flag.Bool("async", false, "use async test")
	link := flag.String("link", "", "link to test")
	mode := flag.String("mode", "pingonly", "speed test mode")
	flag.Parse()
	// link := "vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0"
	if len(*link) < 1 {
		log.Fatal("link required")
	}
	opts := web.ProfileTestOptions{
		GroupName:     "Default",
		SpeedTestMode: *mode,        //  pingonly speedonly all
		PingMethod:    "googleping", // googleping
		SortMethod:    "rspeed",     // speed rspeed ping rping
		Concurrency:   2,
		TestMode:      2, // 2: ALLTEST 3: RETEST
		Subscription:  *link,
		Language:      "en", // en cn
		FontSize:      24,
		Theme:         "rainbow",
		Timeout:       10 * time.Second,
		OutputMode:    0, // 0: base64 1:file path 2: no pic 3: json 4: txt
	}
	ctx := context.Background()
	var err error
	if *async {
		err = pingAsync(ctx, opts)
	} else {
		err = pingSync(ctx, opts)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func pingSync(ctx context.Context, opts web.ProfileTestOptions) error {
	nodes, err := web.TestContext(ctx, opts, &web.EmptyMessageWriter{})
	if err != nil {
		return err
	}

	for _, node := range nodes {
		// tested node info here
		if node.IsOk {
			fmt.Println("id:", node.Id, node.Remarks, "ping:", node.Ping)
		}
	}
	return nil
}

func pingAsync(ctx context.Context, opts web.ProfileTestOptions) error {
	nodeChan, links, err := web.TestAsyncContext(ctx, opts)
	if err != nil {
		return err
	}
	count := len(links)
	for i := 0; i < count; i++ {
		node := <-nodeChan
		if node.IsOk {
			fmt.Println("id:", node.Id, node.Remarks, "ping:", node.Ping, "avg", download.ByteCountIECTrim(node.AvgSpeed), "max", download.ByteCountIECTrim(node.MaxSpeed), "link:", links[node.Id])
		}
	}
	close(nodeChan)
	return nil
}
