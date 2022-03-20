package download

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/common/pool"
)

var (
	contentLength = 242743296
)

func DownloadRange(ctx context.Context, link string, part int, timeout time.Duration, handshakeTimeout time.Duration, resultChan chan<- int64, startChan chan<- time.Time) (int64, error) {
	client, err := createClient(ctx, link)
	if err != nil {
		return 0, err
	}

	option := DownloadOption{
		DownloadTimeout:  timeout,
		HandshakeTimeout: handshakeTimeout,
		URL:              downloadLink,
		Ranges:           calcRange(int64(part), int64(contentLength), link),
	}
	return downloadRangeInternal(ctx, option, resultChan, startChan, client.Dial)
}

func downloadRangeInternal(ctx context.Context, option DownloadOption, resultChan chan<- int64, startOuterChan chan<- time.Time, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	var max int64 = 0
	var wg sync.WaitGroup
	totalCh := make(chan int64)
	// remove
	errorCh := make(chan error)
	startCh := make(chan time.Time, 1)
	for _, rng := range option.Ranges {
		wg.Add(1)
		go func(rng Range, totalChan chan<- int64, errorChan chan<- error, startChan chan<- time.Time) (int64, error) {
			defer wg.Done()
			var max int64 = 0
			httpTransport := &http.Transport{}
			httpClient := &http.Client{Transport: httpTransport, Timeout: option.HandshakeTimeout}
			if dial != nil {
				httpTransport.Dial = dial
			}
			req, err := http.NewRequest("GET", option.URL, nil)
			if err != nil {
				errorChan <- err
				return max, err
			}
			// add range
			ranges := rng.toHeader(int64(contentLength))
			req.Header.Add("Range", ranges)
			response, err := httpClient.Do(req)
			if err != nil {
				errorChan <- err
				return max, err
			}
			defer response.Body.Close()
			prev := time.Now()
			startChan <- prev
			var total int64
			buf := pool.Get(20 * 1024)
			pool.Put(buf)
		Loop:
			for {
				if ctx.Err() != nil {
					break
				}
				select {
				case <-ctx.Done():
					return max, err
				default:
					// buf := pool.Get(20 * 1024)
					nr, er := response.Body.Read(buf)
					total += int64(nr)
					// pool.Put(buf)
					now := time.Now()
					if now.Sub(prev) >= 100*time.Millisecond || er != nil {
						prev = now
						if totalChan != nil {
							totalChan <- total
						}
						if max < total {
							max = total
						}
						total = 0
					}
					if er != nil {
						if er != io.EOF {
							errorChan <- err
							err = er
						}
						break Loop
					}
				}

			}
			return max, nil

		}(rng, totalCh, errorCh, startCh)
	}
	var sum int64 = 0
	var errorResult error = nil

	doneCh := make(chan bool, 1)
	go func(doneChan chan<- bool) {
		wg.Wait()
		doneChan <- true
	}(doneCh)
	var prev time.Time
	for {
		if !prev.IsZero() {
			now := time.Now()
			if now.Sub(prev) >= time.Second {
				prev = now
				if resultChan != nil {
					resultChan <- sum
				}
				if max < sum {
					max = sum
				}
				sum = 0
			}
		}
		select {
		case total := <-totalCh:
			if total < 0 {
				return max, nil
			}
			sum += total
		case err := <-errorCh:
			if err != nil {
				errorResult = err
			}
		case start := <-startCh:
			// init only once
			if prev.IsZero() {
				prev = start
				if startOuterChan != nil {
					startOuterChan <- start
				}
			}
		case <-doneCh:
			return max, errorResult
		case <-ctx.Done():
			err := ctx.Err()
			if max > 0 {
				err = nil
			}
			return max, err
		}
	}
}
