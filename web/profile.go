package web

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/download"
	"github.com/xxf098/lite-proxy/request"
	"github.com/xxf098/lite-proxy/utils"
	"github.com/xxf098/lite-proxy/web/render"
)

var ErrInvalidData = errors.New("invalid data")

const (
	PIC_BASE64 = iota
	PIC_PATH
	PIC_NONE
)

// support proxy
// concurrency setting
// as subscription server
// profiles filter
// clash to vmess local subscription
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
	msg, err := utils.DecodeB64(string(data))
	if err != nil {
		return parseClash(string(data))
	}
	return parseLinks(msg)
}

type parseFunc func(string) ([]string, error)

func parseLinks(message string) ([]string, error) {
	// splits := strings.SplitN(string(message), "^", 2)
	// if len(splits) < 1 {
	// 	return nil, errors.New("Invalid Data")
	// }
	matched, err := regexp.MatchString(`^(?:https?:\/\/)(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`, message)
	if matched && err == nil {
		return getSubscriptionLinks(message)
	}
	var links []string
	for _, fn := range []parseFunc{parseProfiles, parseBase64, parseClash, parseFile} {
		links, err = fn(message)
		if err == nil && len(links) > 0 {
			break
		}
	}
	return links, err
}

func parseProfiles(data string) ([]string, error) {
	reg := regexp.MustCompile(`((?i)(vmess|ssr)://[a-zA-Z0-9+_/=-]+)|((?i)(ss|trojan)://(.+?)@(.+?):([0-9]{2,5})([?#][^\s]+))|((?i)(ss)://[a-zA-Z0-9+_/=-]+([?#][^\s]+))`)
	matches := reg.FindAllStringSubmatch(data, -1)
	links := make([]string, len(matches))
	for index, match := range matches {
		links[index] = match[0]
	}
	return links, nil
}

func parseBase64(data string) ([]string, error) {
	msg, err := utils.DecodeB64(data)
	if err != nil {
		return nil, err
	}
	return parseProfiles(msg)
}

func parseClash(data string) ([]string, error) {
	cc, err := config.ParseClash([]byte(data))
	if err != nil {
		return nil, err
	}
	return cc.Proxies, nil
}

func parseClashByLine(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	data := []byte{}
	checkProfile := false
	for scanner.Scan() {
		b := scanner.Bytes()
		trimLine := strings.TrimSpace(string(b))
		if trimLine == "proxy-groups:" || trimLine == "rules:" {
			break
		}
		if checkProfile {
			if _, err := config.ParseBaseProxy(trimLine); err != nil {
				continue
			}
		}
		// check profile
		if !checkProfile && trimLine == "proxies:" {
			checkProfile = true
		}
		data = append(data, b...)
		data = append(data, byte('\n'))
	}
	return parseClashByte(data)
}

func parseClashByte(data []byte) ([]string, error) {
	cc, err := config.ParseClash(data)
	if err != nil {
		return nil, err
	}
	return cc.Proxies, nil
}

func parseFile(filepath string) ([]string, error) {
	if _, err := os.Stat(filepath); err != nil {
		return nil, err
	}
	// clash
	if strings.HasSuffix(filepath, ".yaml") {
		return parseClashByLine(filepath)
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return parseBase64(string(data))
}

func parseOptions(message string) (*ProfileTestOptions, error) {
	opts := strings.Split(message, "^")
	if len(opts) < 7 {
		return nil, ErrInvalidData
	}
	groupName := opts[0]
	if groupName == "?empty?" || groupName == "" {
		groupName = "Default"
	}
	concurrency, err := strconv.Atoi(opts[5])
	if err != nil {
		return nil, err
	}
	if concurrency < 1 {
		concurrency = 1
	}
	timeout, err := strconv.Atoi(opts[6])
	if err != nil {
		return nil, err
	}
	if timeout < 20 {
		timeout = 20
	}
	testOpt := &ProfileTestOptions{
		GroupName:     groupName,
		SpeedTestMode: opts[1],
		PingMethod:    opts[2],
		SortMethod:    opts[3],
		Concurrency:   concurrency,
		TestMode:      ALLTEST,
		Timeout:       time.Duration(timeout) * time.Second,
	}
	return testOpt, nil
}

const (
	SpeedOnly = "speedonly"
	PingOnly  = "pingonly"
	ALLTEST   = iota
	RETEST
)

type ProfileTestOptions struct {
	GroupName       string        `json:"group"`
	SpeedTestMode   string        `json:"speedtestMode"` // speedonly pingonly all
	PingMethod      string        `json:"pingMethod"`    // googleping
	SortMethod      string        `json:"sortMethod"`    // speed rspeed ping rping
	Concurrency     int           `json:"concurrency"`
	TestMode        int           `json:"testMode"` // 2: ALLTEST 3: RETEST
	TestIDs         []int         `json:"testids"`
	Timeout         time.Duration `json:"timeout"`
	Links           []string      `json:"links"`
	Subscription    string        `json:"subscription"`
	Language        string        `json:"language"`
	FontSize        int           `json:"fontSize"`
	Theme           string        `json:"theme"`
	GeneratePicMode int           `json:"generatePicMode"` // 0: base64 1:file path 2: no pic
}

func parseMessage(message []byte) ([]string, *ProfileTestOptions, error) {
	options := &ProfileTestOptions{}
	err := json.Unmarshal(message, options)
	if err != nil {
		return nil, nil, err
	}
	options.Timeout = time.Duration(int(options.Timeout)) * time.Second
	if options.GroupName == "?empty?" || options.GroupName == "" {
		options.GroupName = "Default"
	}
	if options.Timeout < 8 {
		options.Timeout = 8
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	if options.TestMode == RETEST {
		return options.Links, options, nil
	}
	options.TestMode = ALLTEST
	links, err := parseLinks(options.Subscription)
	if err != nil {
		return nil, nil, err
	}
	return links, options, nil
}

func parseRetestMessage(message []byte) ([]string, *ProfileTestOptions, error) {
	options := &ProfileTestOptions{}
	err := json.Unmarshal(message, options)
	if err != nil {
		return nil, nil, err
	}
	if options.TestMode != RETEST {
		return nil, nil, errors.New("not retest mode")
	}
	options.TestMode = RETEST
	options.Timeout = time.Duration(int(options.Timeout)) * time.Second
	if options.GroupName == "?empty?" || options.GroupName == "" {
		options.GroupName = "Default"
	}
	if options.Timeout < 20 {
		options.Timeout = 20
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	return options.Links, options, nil
}

type MessageWriter interface {
	WriteMessage(messageType int, data []byte) error
}

type OutputMessageWriter struct {
}

func (p *OutputMessageWriter) WriteMessage(messageType int, data []byte) error {
	log.Println(string(data))
	return nil
}

type EmptyMessageWriter struct {
}

func (w *EmptyMessageWriter) WriteMessage(messageType int, data []byte) error {
	return nil
}

type ProfileTest struct {
	Writer      MessageWriter
	Options     *ProfileTestOptions
	MessageType int
	Links       []string
	mu          sync.Mutex
	wg          sync.WaitGroup // wait for all to finish
}

func (p *ProfileTest) WriteMessage(data []byte) error {
	var err error
	if p.Writer != nil {
		p.mu.Lock()
		err = p.Writer.WriteMessage(p.MessageType, data)
		p.mu.Unlock()
	}
	return err
}

func (p *ProfileTest) WriteString(data string) error {
	b := []byte(data)
	return p.WriteMessage(b)
}

// api
func (p *ProfileTest) TestAll(ctx context.Context, links []string, max int, trafficChan chan<- int64) (chan render.Node, error) {
	linksCount := len(links)
	if linksCount < 1 {
		return nil, fmt.Errorf("no profile found")
	}

	nodeChan := make(chan render.Node, linksCount)
	go func(context.Context) {
		guard := make(chan int, max)
		for i := range links {
			p.wg.Add(1)
			id := i
			link := links[i]
			select {
			case guard <- i:
				go func(id int, link string, c <-chan int, nodeChan chan<- render.Node) {
					p.testOne(ctx, id, link, nodeChan, trafficChan)
					<-c
				}(id, link, guard, nodeChan)
			case <-ctx.Done():
				return
			}
		}
		p.wg.Wait()
		// if trafficChan != nil {
		// 	close(trafficChan)
		// }
	}(ctx)
	return nodeChan, nil
}

func (p *ProfileTest) testAll(ctx context.Context) (render.Nodes, error) {
	linksCount := len(p.Links)
	if linksCount < 1 {
		p.WriteString(SPEEDTEST_ERROR_NONODES)
		return nil, fmt.Errorf("no profile found")
	}
	start := time.Now()
	p.WriteMessage(getMsgByte(-1, "started"))
	for i := range p.Links {
		p.WriteMessage(gotserverMsg(i, p.Links[i], p.Options.GroupName))
	}
	guard := make(chan int, p.Options.Concurrency)
	nodeChan := make(chan render.Node, linksCount)

	nodes := make(render.Nodes, linksCount)
	for i := range p.Links {
		p.wg.Add(1)
		id := i
		link := ""
		if len(p.Options.TestIDs) > 0 && len(p.Options.Links) > 0 {
			id = p.Options.TestIDs[i]
			link = p.Options.Links[i]
		}
		select {
		case guard <- i:
			go func(id int, link string, c <-chan int, nodeChan chan<- render.Node) {
				p.testOne(ctx, id, link, nodeChan, nil)
				_ = p.WriteMessage(getMsgByte(id, "endone"))
				<-c
			}(id, link, guard, nodeChan)
		case <-ctx.Done():
			return nil, nil
		}
	}
	p.wg.Wait()
	p.WriteMessage(getMsgByte(-1, "eof"))
	duration := FormatDuration(time.Since(start))
	// draw png
	successCount := 0
	var traffic int64 = 0
	for i := 0; i < linksCount; i++ {
		node := <-nodeChan
		nodes[node.Id] = node
		traffic += node.Traffic
		if node.IsOk {
			successCount += 1
		}
	}
	close(nodeChan)

	if p.Options.GeneratePicMode == PIC_NONE {
		return nodes, nil
	}

	// sort nodes
	nodes.Sort(p.Options.SortMethod)

	fontPath := "WenQuanYiMicroHei-01.ttf"
	options := render.NewTableOptions(40, 30, 0.5, 0.5, p.Options.FontSize, 0.5, fontPath, p.Options.Language, p.Options.Theme, "Asia/Shanghai", FontBytes)
	table, err := render.NewTableWithOption(nodes, &options)
	if err != nil {
		return nil, err
	}
	// msg := fmt.Sprintf("Total Traffic : %s. Total Time : %s. Working Nodes: [%d/%d]", download.ByteCountIECTrim(traffic), duration, successCount, linksCount)
	msg := table.FormatTraffic(download.ByteCountIECTrim(traffic), duration, fmt.Sprintf("%d/%d", successCount, linksCount))
	if p.Options.GeneratePicMode == PIC_PATH {
		table.Draw("out.png", msg)
		p.WriteMessage(getMsgByte(-1, "picdata", "out.png"))
		return nodes, nil
	}
	if picdata, err := table.EncodeB64(msg); err == nil {
		p.WriteMessage(getMsgByte(-1, "picdata", picdata))
	}
	return nodes, nil
}

func (p *ProfileTest) testOne(ctx context.Context, index int, link string, nodeChan chan<- render.Node, trafficChan chan<- int64) error {
	// panic
	defer p.wg.Done()
	if link == "" {
		link = p.Links[index]
		link = strings.SplitN(link, "^", 2)[0]
	}
	protocol, remarks, err := GetRemarks(link)
	if err != nil || remarks == "" {
		remarks = fmt.Sprintf("Profile %d", index)
	}
	elapse, err := p.pingLink(index, link)
	if err != nil {
		node := render.Node{
			Id:       index,
			Group:    p.Options.GroupName,
			Remarks:  remarks,
			Protocol: protocol,
			Ping:     fmt.Sprintf("%d", elapse),
			AvgSpeed: 0,
			MaxSpeed: 0,
			IsOk:     elapse > 0,
		}
		nodeChan <- node
		return err
	}
	err = p.WriteMessage(getMsgByte(index, "startspeed"))
	ch := make(chan int64, 1)
	startCh := make(chan time.Time, 1)
	defer close(ch)
	go func(ch <-chan int64, startChan <-chan time.Time) {
		var max int64
		var sum int64
		var avg int64
		start := time.Now()
	Loop:
		for {
			select {
			case speed, ok := <-ch:
				if !ok || speed < 0 {
					break Loop
				}
				sum += speed
				duration := float64(time.Since(start)/time.Millisecond) / float64(1000)
				avg = int64(float64(sum) / duration)
				if max < speed {
					max = speed
				}
				log.Printf("%s recv: %s", remarks, download.ByteCountIEC(speed))
				err = p.WriteMessage(getMsgByte(index, "gotspeed", avg, max, speed))
				if trafficChan != nil {
					trafficChan <- speed
				}
			case s := <-startChan:
				start = s
			case <-ctx.Done():
				log.Printf("index %d done!", index)
				break Loop
			}
		}
		node := render.Node{
			Id:       index,
			Group:    p.Options.GroupName,
			Remarks:  remarks,
			Protocol: protocol,
			Ping:     fmt.Sprintf("%d", elapse),
			AvgSpeed: avg,
			MaxSpeed: max,
			IsOk:     true,
			Traffic:  sum,
		}
		nodeChan <- node
	}(ch, startCh)
	speed, err := download.Download(link, p.Options.Timeout, p.Options.Timeout, ch, startCh)
	// speed, err := download.DownloadRange(link, 2, p.Options.Timeout, p.Options.Timeout, ch, startCh)
	if speed < 1 {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
	}
	return err
}

func (p *ProfileTest) pingLink(index int, link string) (int64, error) {
	if p.Options.SpeedTestMode == SpeedOnly {
		return 0, nil
	}
	if link == "" {
		link = p.Links[index]
	}
	p.WriteMessage(getMsgByte(index, "startping"))
	elapse, err := request.PingLink(link, 2)
	p.WriteMessage(getMsgByte(index, "gotping", elapse))
	if elapse < 1 {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
		return 0, err
	}
	if p.Options.SpeedTestMode == PingOnly {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
		return elapse, errors.New(PingOnly)
	}
	return elapse, err
}

func FormatDuration(duration time.Duration) string {
	h := duration / time.Hour
	duration -= h * time.Hour
	m := duration / time.Minute
	duration -= m * time.Minute
	s := duration / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

func png2base64(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes), nil
}
