package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/download"
	"github.com/xxf098/lite-proxy/request"
)

// support proxy
func getSubscriptionLinks(link string) ([]string, error) {
	c := http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := c.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	msg, err := common.DecodeB64Bytes(string(data))
	if err != nil {
		return nil, err
	}
	return parseLinks(msg)
}

func parseLinks(message []byte) ([]string, error) {
	splits := strings.SplitN(string(message), "^", 2)
	if len(splits) < 1 {
		return nil, errors.New("Invalid Data")
	}
	matched, err := regexp.MatchString(`^(?:https?:\/\/)(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`, splits[0])
	if matched && err == nil {
		return getSubscriptionLinks(splits[0])
	}
	reg := regexp.MustCompile("(?i)(vmess|ssr|ss|trojan)://[a-zA-Z0-9+_/=-]+")
	matches := reg.FindAllStringSubmatch(splits[0], -1)
	links := make([]string, len(matches))
	for index, match := range matches {
		links[index] = match[0]
	}
	return links, nil
}

type ProfileTestOptions struct {
	Concurrency int
	Timeout     time.Duration
}

type ProfileTest struct {
	Conn        *websocket.Conn
	Options     ProfileTestOptions
	MessageType int
	Links       []string
	mu          sync.Mutex
	wg          sync.WaitGroup // wait for all to finish
}

func (p *ProfileTest) WriteMessage(data []byte) error {
	p.mu.Lock()
	err := p.Conn.WriteMessage(p.MessageType, data)
	p.mu.Unlock()
	return err
}

func (p *ProfileTest) WriteString(data string) error {
	b := []byte(data)
	return p.WriteMessage(b)
}

func (p *ProfileTest) testAll(ctx context.Context) error {
	if len(p.Links) < 1 {
		p.WriteString(SPEEDTEST_ERROR_NONODES)
		return fmt.Errorf("No nodes found!")
	}
	p.WriteMessage(getMsgByte(-1, "started"))
	for i, _ := range p.Links {
		p.WriteMessage(gotserverMsg(i, p.Links[i]))
	}
	guard := make(chan int, p.Options.Concurrency)
	for i, _ := range p.Links {
		p.wg.Add(1)
		select {
		case guard <- i:
			go func(index int, c <-chan int) {
				p.testSingle(ctx, index)
				<-c
			}(i, guard)
		case <-ctx.Done():
			break
		}
	}
	p.wg.Wait()
	p.WriteMessage(getMsgByte(-1, "eof"))
	return nil
}

func (p *ProfileTest) testSingle(ctx context.Context, index int) error {
	// panic
	defer p.wg.Done()
	p.WriteMessage(getMsgByte(index, "startping"))
	link := p.Links[index]
	link = strings.SplitN(link, "^", 2)[0]
	elapse, err := request.PingLink(link)
	err = p.WriteMessage(getMsgByte(index, "gotping", elapse))
	if elapse < 1 {
		return err
	}
	err = p.WriteMessage(getMsgByte(index, "startspeed"))
	ch := make(chan int64, 1)
	go func(ch <-chan int64) {
		var max int64
		var speeds []int64
		for {
			select {
			case speed := <-ch:
				if speed < 0 {
					return
				}
				speeds = append(speeds, speed)
				var avg int64
				for _, s := range speeds {
					avg += s / int64(len(speeds))
				}
				if max < speed {
					max = speed
				}
				log.Printf("recv: %s", download.ByteCountIEC(speed))
				err = p.WriteMessage(getMsgByte(index, "gotspeed", avg, max))
			case <-ctx.Done():
				log.Printf("index %d done!", index)
				return
			}
		}
	}(ch)
	download.Download(link, p.Options.Timeout, p.Options.Timeout, ch)
	return err
}

type Message struct {
	ID       int    `json:"id"`
	Info     string `json:"info"`
	Remarks  string `json:"remarks"`
	Group    string `json:"group"`
	Ping     int64  `json:"ping"`
	Lost     string `json:"lost"`
	Speed    string `json:"speed"`
	MaxSpeed string `json:"maxspeed"`
}

func gotserverMsg(id int, link string) []byte {
	msg := Message{ID: id, Info: "gotserver"}
	cfg, err := config.VmessLinkToVmessConfigIP(link, false)
	if err == nil {
		msg.Group = "Group 1"
		msg.Remarks = cfg.Ps
	}
	b, _ := json.Marshal(msg)
	return b
}

func getMsgByte(id int, typ string, option ...interface{}) []byte {
	msg := Message{ID: id, Info: typ}
	switch typ {
	case "gotserver":
		msg.Remarks = "Server 1"
		msg.Group = "Group 1"
	case "gotping":
		msg.Lost = "0.00%"
		var ping int64
		if len(option) > 0 {
			if v, ok := option[0].(int64); ok {
				ping = v
			}
		}
		msg.Ping = ping
	case "gotspeed":
		var speed int64
		var maxspeed int64
		if len(option) > 1 {
			if v, ok := option[0].(int64); ok {
				speed = v
			}
			if v, ok := option[1].(int64); ok {
				maxspeed = v
			}
		}
		msg.Speed = strings.TrimRight(download.ByteCountIEC(speed), "/s")
		msg.MaxSpeed = strings.TrimRight(download.ByteCountIEC(maxspeed), "/s")
	}
	b, _ := json.Marshal(msg)
	return b
}
