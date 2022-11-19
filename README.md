# LiteSpeedTest

LiteSpeedTest is a simple tool for batch test ss/ssr/v2ray/trojan/clash servers.   
Support tested by single link, multiple links, subscription link and file path.

 ![build](https://github.com/xxf098/LiteSpeedTest/actions/workflows/test.yaml/badge.svg?branch=master&event=push) 

### Usage
```
Run as a speed test tool:
    # run this command then open http://127.0.0.1:10888/ in your browser to start speed test
    ./lite
    # start with another port
    ./lite -p 10889
    
    # test in command line only mode
    ./lite --test https://raw.githubusercontent.com/freefq/free/master/v2
    # test in command line only mode with custom config.
    # details can find here https://github.com/xxf098/LiteSpeedTest/blob/master/config.json
    # all config options:
    #       "group":"job",   // group name
	#       "speedtestMode":"pingonly", // speedonly pingonly all
	#       "pingMethod":"googleping",  // googleping tcpping
	#       "sortMethod":"rspeed",      // speed rspeed ping rping
	#       "concurrency":1,  // concurrency number
	#       "testMode":2,   // 2: ALLTEST 3: RETEST
	#       "subscription":"subscription url",
	#       "timeout":16,  // timeout in seconds
	#       "language":"en", // en cn
	#       "fontSize":24,
	#       "unique": true,  // remove duplicated value
	#       "theme":"rainbow", 
	#       "generatePicMode": 1  // 0: base64 1: pic path 2: no pic 3: json
    ./lite --config config.json --test https://raw.githubusercontent.com/freefq/free/master/v2


Run as a grpc server:
    # start the grpc server  
    ./lite -grpc -p 10999
    # grpc go client example in ./api/rpc/liteclient/client.go 
    # grpc python client example in ./api/rpc/liteclientpy/client.py

Run as a http/socks5 proxy:
    # use default port 8090
    ./lite vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
    ./lite ssr://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
    # use another port
    ./lite -p 8091 vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
```

### Build
```bash
    #require go>=1.18.1
    GOOS=js GOARCH=wasm go get -u ./...
    cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./web/wasm_exec.js
    GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./wasm
    go build -o lite
```

### Docker
```bash
 docker build --network=host  -t lite:$(git describe --tags --abbrev=0) -f ./docker/Dockerfile ./
 docker run -p 10888:10888/tcp lite:$(git describe --tags --abbrev=0)
```

## Credits

- [clash](https://github.com/Dreamacro/clash)
- [stairspeedtest-reborn](https://github.com/tindy2013/stairspeedtest-reborn)
- [gg](https://github.com/fogleman/gg)

## Developer
```golang
import (
    "context"
    "fmt"
	"time"
    "github.com/xxf098/lite-proxy/web"
)
// see more details in ./examples
func testPing() error {
    ctx := context.Background()
    link := "https://www.example.com/subscription/link"
    opts := web.ProfileTestOptions{
		GroupName:     "Default", 
		SpeedTestMode: "pingonly",   //  pingonly speedonly all
		PingMethod:    "googleping", // googleping
		SortMethod:    "rspeed", // speed rspeed ping rping
		Concurrency:   2,
		TestMode:      2,
		Subscription:  link,
		Language:      "en",  // en cn
		FontSize:      24,
		Theme:         "rainbow",
        Unique:        true,
		Timeout:       10 * time.Second,
		GeneratePicMode:  0,
	}
    nodes, err := web.TestContext(ctx, opts, &web.EmptyMessageWriter{})
    if err != nil {
        return err
    }
    // get all ok profile
    for _, node := range nodes {
        if node.IsOk {
			fmt.Println(node.Remarks)
		}
	}
    return nil
}
```
