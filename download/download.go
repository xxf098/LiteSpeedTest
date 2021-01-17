package download

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"errors"
	"regexp"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/common/pool"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/stats"
)

const (
	downloadLink = "https://download.microsoft.com/download/2/0/E/20E90413-712F-438C-988E-FDAA79A8AC3D/dotnetfx35.exe"
)

type Empty struct {
	total stats.Counter
}

func (e *Empty) Write(p []byte) (n int, err error) {
	n = len(p)
	pool.Put(p)
	e.total.Add(int64(n))
	// fmt.Printf("==%s\n", byteCountIEC(int64(n)))
	return n, nil
}

func (e *Empty) Size() int64 {
	return e.total.Set(0)
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B/S", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB/S",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func createClient(ctx context.Context, link string) (*proxy.Client, error) {
	var d outbound.Dialer
	r := regexp.MustCompile("(?i)^(vmess|trojan|ss|ssr)://.+")
	matches := r.FindStringSubmatch(link)
	if len(matches) < 2 {
		return nil, common.NewError("Not Suported Link")
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

func Download(link string, timeout time.Duration, resultChan chan<- int64) (int64, error) {
	ctx := context.Background()
	client, err := createClient(ctx, link)
	if err != nil {
		return 0, err
	}
	return downloadInternal(ctx, downloadLink, timeout, resultChan, client.Dial)
}

func downloadInternal(ctx context.Context, url string, timeout time.Duration, resultChan chan<- int64, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	var max int64 = 0
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
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
	output := &Empty{
		total: stats.Counter{},
	}
	go func(response *http.Response) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				size := output.Size()
				if max < size {
					max = size
				}
				if resultChan != nil {
					resultChan <- size
				}
				// fmt.Printf("%s\n", byteCountIEC(size))
			case <-ctx.Done():
				if resultChan != nil {
					resultChan <- -1
				}
				response.Body.Close()
				// fmt.Println("Done")
				return
			}
		}
	}(response)

	_, err = common.CopyBuffer(output, response.Body, pool.Get(20*1024))
	// if err != nil {
	// 	return max, err
	// }
	return max, nil
}
