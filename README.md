# LiteSpeedTest

LiteSpeedTest is a simple tool for batch test ss/ssr/v2ray/trojan servers. 

### Usage
```
Run as speed test tool:
    ./lite    

Run as http/socks5 proxy:
    ./lite vmess://ABCDEFGHIJKLMNOPQRSTUVWXYZ
    ./lite ssr://ABCDEFGHIJKLMNOPQRSTUVWXYZ
```

### Build
```bash
    go get -u ./...
    # packr2
    # export GO111MODULE=on
    go get -u github.com/gobuffalo/packr/v2/...
    packr2
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
