# LiteSpeedTest

LiteSpeedTest is a simple tool for batch test ss/ssr/v2ray/trojan server. 

### Usage
```
As proxy:
    lite vmess://ABCDEFGHIJKLMNOPQRSTUVWXYZ
    lite ssr://ABCDEFGHIJKLMNOPQRSTUVWXYZ

As test tool:
    lite
```

### Build
```bash
    go get -u ./...
    # packr2
    go get -u github.com/gobuffalo/packr/v2/...
    packr2
    go build -o lite
```

## Developer
```golang
// release
```
