package download

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"errors"

	"github.com/xxf098/lite-proxy/common/pool"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/stats"
	"github.com/xxf098/lite-proxy/utils"
)

const (
	downloadLink      = "https://download.microsoft.com/download/2/0/E/20E90413-712F-438C-988E-FDAA79A8AC3D/dotnetfx35.exe"
	cloudflareLink100 = "https://speed.cloudflare.com/__down?bytes=100000000"
	cachefly10        = "http://cachefly.cachefly.net/10mb.test"
	cachefly100       = "http://cachefly.cachefly.net/100mb.test"
)

type DownloadOption struct {
	URL              string
	DownloadTimeout  time.Duration
	HandshakeTimeout time.Duration
	Ranges           []Range
}

type Discard struct {
	total stats.Counter
}

func (e *Discard) Write(p []byte) (n int, err error) {
	n = len(p)
	pool.Put(p)
	e.total.Add(int64(n))
	// fmt.Printf("==%s\n", ByteCountIEC(int64(n)))
	return n, nil
}

func (e *Discard) Size() int64 {
	return e.total.Set(0)
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B/s", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB/s",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func ByteCountIECTrim(b int64) string {
	result := ByteCountIEC(b)
	return strings.TrimSuffix(result, "/s")
}

func createClient(ctx context.Context, link string) (*proxy.Client, error) {
	var d outbound.Dialer
	matches, err := utils.CheckLink(link)
	if err != nil {
		return nil, err
	}
	creator, err := outbound.GetDialerCreator(matches[1])
	if err != nil {
		return nil, err
	}
	d, err = creator(link)
	if err != nil {
		return nil, err
	}
	if d != nil {
		return proxy.NewClient(ctx, d), nil
	}

	return nil, errors.New("not supported link")
}

func Download(link string, timeout time.Duration, handshakeTimeout time.Duration, resultChan chan<- int64, startChan chan<- time.Time) (int64, error) {
	ctx := context.Background()
	client, err := createClient(ctx, link)
	if err != nil {
		return 0, err
	}
	option := DownloadOption{
		DownloadTimeout:  timeout,
		HandshakeTimeout: handshakeTimeout,
		URL:              downloadLink,
	}
	return downloadInternal(ctx, option, resultChan, startChan, client.Dial)
}

func downloadInternal(ctx context.Context, option DownloadOption, resultChan chan<- int64, startOuterChan chan<- time.Time, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	var max int64 = 0
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport, Timeout: option.HandshakeTimeout}
	if dial != nil {
		httpTransport.Dial = dial
	}
	req, err := http.NewRequest("GET", option.URL, nil)
	if err != nil {
		return max, err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return max, err
	}
	defer response.Body.Close()
	prev := time.Now()
	if startOuterChan != nil {
		startOuterChan <- prev
	}
	var total int64
	buf := pool.Get(20 * 1024)
	defer pool.Put(buf)
	for {
		// buf := pool.Get(20 * 1024)
		nr, er := response.Body.Read(buf)
		total += int64(nr)
		// pool.Put(buf)
		now := time.Now()
		if now.Sub(prev) >= time.Second || er != nil {
			prev = now
			if resultChan != nil {
				resultChan <- total
			}
			if max < total {
				max = total
			}
			total = 0
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return max, nil
}

func DownloadComplete(link string, timeout time.Duration, handshakeTimeout time.Duration) (int64, error) {
	ctx := context.Background()
	client, err := createClient(ctx, link)
	if err != nil {
		return 0, err
	}
	return downloadCompleteInternal(ctx, cachefly100, timeout, handshakeTimeout, client.Dial)
}

func downloadCompleteInternal(ctx context.Context, url string, timeout time.Duration, handshakeTimeout time.Duration, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	var max int64 = 0
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport, Timeout: handshakeTimeout}
	if dial != nil {
		httpTransport.Dial = dial
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return max, err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return max, err
	}
	defer response.Body.Close()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	start := time.Now()
	var total int64
	buf := pool.Get(20 * 1024)
	defer pool.Put(buf)
	for ctx.Err() == nil {
		nr, er := response.Body.Read(buf)
		total += int64(nr)
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	now := time.Now()
	max = total * 1000 / now.Sub(start).Milliseconds()
	return max, nil
}

func WSDownload(link string, timeout time.Duration, handshakeTimeout time.Duration, resultChan chan<- int64) (int64, error) {
	return 0, nil
}
