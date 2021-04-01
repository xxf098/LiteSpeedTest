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
    # go-bindata
    go get -u github.com/go-bindata/go-bindata/...
    go-bindata -fs -pkg web -prefix "web/gui"  -o ./web/asset.go web/gui/
    go build -o lite
```

## Developer
```golang
// release
```
