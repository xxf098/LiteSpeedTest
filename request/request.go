package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/outbound"
)

func PingVmess(vmessOption *outbound.VmessOption) (int64, error) {
	tcpTimeout := 2100 * time.Second
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
	if err != nil {
		return 0, err
	}
	defer remoteConn.Close()
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
	_, err = parseFirstLine(buf)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(start).Milliseconds()
	// fmt.Print(string(buf))
	// fmt.Printf("server: %s port: %d elapsed: %d\n", vmessOption.Server, vmessOption.Port, elapsed)
	return elapsed, nil
}

func parseFirstLine(buf []byte) (int, error) {
	bNext := buf
	var b []byte
	var err error
	for len(b) == 0 {
		if b, bNext, err = nextLine(bNext); err != nil {
			return 0, err
		}
	}

	// parse protocol
	n := bytes.IndexByte(b, ' ')
	if n < 0 {
		return 0, fmt.Errorf("cannot find whitespace in the first line of response %q", buf)
	}
	b = b[n+1:]

	// parse status code
	statusCode, n, err := parseUintBuf(b)
	if err != nil {
		return 0, fmt.Errorf("cannot parse response status code: %s. Response %q", err, buf)
	}
	if len(b) > n && b[n] != ' ' {
		return 0, fmt.Errorf("unexpected char at the end of status code. Response %q", buf)
	}

	if statusCode == 204 || statusCode == 200 {
		return len(buf) - len(bNext), nil
	}
	return 0, errors.New("Wrong Status Code")
}

func nextLine(b []byte) ([]byte, []byte, error) {
	nNext := bytes.IndexByte(b, '\n')
	if nNext < 0 {
		return nil, nil, errors.New("need more data: cannot find trailing lf")
	}
	n := nNext
	if n > 0 && b[n-1] == '\r' {
		n--
	}
	return b[:n], b[nNext+1:], nil
}

func parseUintBuf(b []byte) (int, int, error) {
	n := len(b)
	if n == 0 {
		return -1, 0, errors.New("empty integer")
	}
	v := 0
	for i := 0; i < n; i++ {
		c := b[i]
		k := c - '0'
		if k > 9 {
			if i == 0 {
				return -1, i, errors.New("unexpected first char found. Expecting 0-9")
			}
			return v, i, nil
		}
		vNew := 10*v + int(k)
		// Test for overflow.
		if vNew < v {
			return -1, i, errors.New("too long int")
		}
		v = vNew
	}
	return v, n, nil
}

func RequestTrojan(trojanOption *outbound.TrojanOption) (int64, error) {
	tcpTimeout := 2100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), tcpTimeout)
	defer cancel()
	trojan, err := outbound.NewTrojan(*trojanOption)
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
	remoteConn, err := trojan.DialContext(ctx, meta)
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
