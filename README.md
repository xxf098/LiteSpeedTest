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
    #require go>=1.16
    go get -u ./...
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
