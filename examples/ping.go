package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xxf098/lite-proxy/web"
)

func main() {
	ctx := context.Background()
	link := "vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0"
	opts := web.ProfileTestOptions{
		GroupName:       "Default",
		SpeedTestMode:   "pingonly",   //  pingonly speedonly all
		PingMethod:      "googleping", // googleping
		SortMethod:      "rspeed",     // speed rspeed ping rping
		Concurrency:     2,
		TestMode:        2, // 0: all 1: speed only 2: ping only
		Subscription:    link,
		Language:        "en", // en cn
		FontSize:        24,
		Theme:           "rainbow",
		Timeout:         10 * time.Second,
		GeneratePicMode: 0, // 0: base64 1:file path 2: no pic
	}
	nodes, err := web.TestContext(ctx, opts, &web.EmptyMessageWriter{})

	if err != nil {
		return
	}

	for _, node := range nodes {
		// process node info here
		if node.IsOk {
			fmt.Println(node.Remarks, node.Ping, node.Link)
		}
	}
}
