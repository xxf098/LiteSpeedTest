package download

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/common/pool"
	"github.com/xxf098/lite-proxy/stats"
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

func byteCountIEC(b int64) string {
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

func downloadInternal(url string, dial func(network, addr string) (net.Conn, error)) (int64, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
				// fmt.Printf("%s\n", byteCountIEC(size))
			case <-ctx.Done():
				response.Body.Close()
				fmt.Println("Done")
				return
			}
		}
	}(response)

	_, err = common.CopyBuffer(output, response.Body, pool.Get(32*1024))
	if err != nil {
		return max, err
	}
	return max, nil
}
