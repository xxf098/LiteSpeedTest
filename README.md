# LiteSpeedTest

LiteSpeedTest is a simple tool for batch test ss/ssr/v2ray/trojan servers. 

 ![build](https://github.com/xxf098/LiteSpeedTest/actions/workflows/test.yaml/badge.svg?branch=master&event=push) 

### Usage
```
Run as speed test tool:
    # run this command then open http://127.0.0.1:10888/ in your browser to start speed test
    ./lite
    ./lite -p 10889
    # test in command line only mode
    ./lite --test https://raw.githubusercontent.com/freefq/free/master/v2
    # test in command only line mode with custom config. details can find here https://github.com/xxf098/LiteSpeedTest/blob/master/config.json
    ./lite --config config.json --test https://raw.githubusercontent.com/freefq/free/master/v2

Run as http/socks5 proxy:
    ./lite vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
    ./lite ssr://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
    ./lite -p 8091 vmess://aHR0cHM6Ly9naXRodWIuY29tL3h4ZjA5OC9MaXRlU3BlZWRUZXN0
```

### Build
```bash
    #require go>=1.16
    GOOS=js GOARCH=wasm go get -u ./...
    GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./wasm
    go build -o lite
```

## Credits

- [clash](https://github.com/Dreamacro/clash)
- [stairspeedtest-reborn](https://github.com/tindy2013/stairspeedtest-reborn)
- [gg](https://github.com/fogleman/gg)

## Developer
```golang
// TODO
```
