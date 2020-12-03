package request

import (
	"context"
	"fmt"
	"io"
	"time"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/outbound"
)

func Request(vmessOption *outbound.VmessOption) (int64, error) {
	tcpTimeout := 4 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), tcpTimeout)
	defer cancel()
	vmess, err := outbound.NewVmess(*vmessOption)
	if err != nil {
		return 0, err
	}
	remoteHost := "clients3.google.com"
	meta := &C.Metadata{
		NetWork:  0,
		Type:     0,
		SrcPort:  "",
		DstPort:  "80",
		AddrType: 3,
		Host:     remoteHost,
	}
	remoteConn, err := vmess.DialContext(ctx, meta)
	defer remoteConn.Close()
	if err != nil {
		return 0, err
	}
	remoteConn.SetDeadline(time.Now().Add(tcpTimeout))
	start := time.Now()
	httpRequest := "GET /generate_204 HTTP/1.1\r\nHost: %s\r\nUser-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36\r\n\r\n"
	if _, err = fmt.Fprintf(remoteConn, httpRequest, remoteHost); err != nil {
		return 0, err
	}
	buf := make([]byte, 25)
	_, err = remoteConn.Read(buf)
	if err != nil && err != io.EOF {
		return 0, err
	}
	elapsed := time.Since(start).Milliseconds()
	fmt.Print(string(buf))
	// fmt.Printf("server: %s port: %d elapsed: %d\n", vmessOption.Server, vmessOption.Port, elapsed)
	return elapsed, nil
}
