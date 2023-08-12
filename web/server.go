package web

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/core"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/utils"
	"github.com/xxf098/lite-proxy/web/render"
)

var (
	upgrader  = websocket.Upgrader{}
	prevProxy *proxy.Proxy
)

func ServeFile(port int) error {
	// TODO: Mobile UI
	http.HandleFunc("/", serverFile)
	http.HandleFunc("/test", updateTest)
	http.HandleFunc("/getSubscriptionLink", getSubscriptionLink)
	http.HandleFunc("/startProxy", startProxy)
	http.HandleFunc("/getSubscription", getSubscription)
	http.HandleFunc("/getUserConfig", getUserConfig)
	http.HandleFunc("/generateResult", generateResult)
	// log.Printf("Start server at http://127.0.0.1:%d\n", port)
	if ipAddrs, err := listAddresses(); err == nil {
		for _, ipAddr := range ipAddrs {
			log.Printf("Start server at http://%s", net.JoinHostPort(ipAddr, strconv.Itoa(port)))
		}
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return err
}

// func ServeWasm(port int) error {
// 	http.Handle("/", http.FileServer(http.FS(wasmStatic)))
// 	log.Printf("Start server at http://127.0.0.1:%d", port)
// 	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
// 	return err
// }

func serverFile(w http.ResponseWriter, r *http.Request) {
	h := http.FileServer(http.FS(guiStatic))
	r.URL.Path = "gui/dist" + r.URL.Path
	h.ServeHTTP(w, r)
}

func updateTest(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// log.Printf("recv: %s", message)
		links, options, err := parseMessage(message)
		if err != nil {
			msg := `{"info": "error", "reason": "invalidsub"}`
			c.WriteMessage(mt, []byte(msg))
			continue
		}
		if options.Unique {
			uniqueLinks := []string{}
			uniqueMap := map[string]struct{}{}
			for _, link := range links {
				cfg, err := config.Link2Config(link)
				if err != nil {
					continue
				}
				key := fmt.Sprintf("%s%d%s%s%s", cfg.Server, cfg.Port, cfg.Password, cfg.Protocol, cfg.SNI)
				if _, ok := uniqueMap[key]; !ok {
					uniqueLinks = append(uniqueLinks, link)
					uniqueMap[key] = struct{}{}
				}
			}
			links = uniqueLinks
		}
		p := ProfileTest{
			Writer:      c,
			MessageType: mt,
			Links:       links,
			Options:     options,
		}
		go p.testAll(ctx)
		saveUserOptions(options, strings.Split(r.RemoteAddr, ":")[0])
		// err = c.WriteMessage(mt, getMsgByte(0, "gotspeed"))
		// if err != nil {
		// 	log.Println("write:", err)
		// 	break
		// }
	}
}

func readConfig(configPath string) (*ProfileTestOptions, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	options := &ProfileTestOptions{}
	if err = json.Unmarshal(data, options); err != nil {
		return nil, err
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	if options.Language == "" {
		options.Language = "en"
	}
	if options.Theme == "" {
		options.Theme = "rainbow"
	}
	if options.Timeout < 8 {
		options.Timeout = 8
	}
	options.Timeout = options.Timeout * time.Second
	return options, nil
}

func TestFromCMD(subscription string, configPath *string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	options := ProfileTestOptions{
		GroupName:       "Default",
		SpeedTestMode:   "all",
		PingMethod:      "googleping",
		SortMethod:      "rspeed",
		Concurrency:     2,
		TestMode:        2,
		Subscription:    subscription,
		Language:        "en",
		FontSize:        24,
		Theme:           "rainbow",
		Timeout:         15 * time.Second,
		GeneratePicMode: PIC_PATH,
		OutputMode:      PIC_PATH,
	}
	if configPath != nil {
		if opt, err := readConfig(*configPath); err == nil {
			options = *opt
			if options.GeneratePicMode != 0 {
				options.OutputMode = options.GeneratePicMode
			}
			// options.GeneratePic = true
		}
	}
	// check url
	if len(subscription) > 0 && subscription != options.Subscription {
		if _, err := url.Parse(subscription); err == nil {
			options.Subscription = subscription
		} else if _, err := os.Stat(subscription); err == nil {
			options.Subscription = subscription
		}
	}
	if jsonOpt, err := json.Marshal(options); err == nil {
		log.Printf("json options: %s\n", string(jsonOpt))
	}
	_, err := TestContext(ctx, options, &OutputMessageWriter{})
	return err
}

// use as golang api
func TestContext(ctx context.Context, options ProfileTestOptions, w MessageWriter) (render.Nodes, error) {
	links, err := ParseLinks(options.Subscription)
	if err != nil {
		return nil, err
	}
	// outputMessageWriter := OutputMessageWriter{}
	p := ProfileTest{
		Writer:      w,
		MessageType: 1,
		Links:       links,
		Options:     &options,
	}
	return p.testAll(ctx)
}

// use as golang api
func TestAsyncContext(ctx context.Context, options ProfileTestOptions) (chan render.Node, []string, error) {
	links, err := ParseLinks(options.Subscription)
	if err != nil {
		return nil, nil, err
	}
	// outputMessageWriter := OutputMessageWriter{}
	p := ProfileTest{
		Writer:      nil,
		MessageType: ALLTEST,
		Links:       links,
		Options:     &options,
	}
	nodeChan, err := p.TestAll(ctx, nil)
	return nodeChan, links, err
}

type TestResult struct {
	TotalTraffic string `json:"totalTraffic"`
	TotalTime    string `json:"totalTime"`
	Language     string `json:"language"`
	FontSize     int    `json:"fontSize"`
	Theme        string `json:"theme"`
	// SortMethod   string       `json:"sortMethod"`
	Nodes render.Nodes `json:"nodes"`
}

func generateResult(w http.ResponseWriter, r *http.Request) {
	result := TestResult{}
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	if err = json.Unmarshal(data, &result); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fontPath := "WenQuanYiMicroHei-01.ttf"
	options := render.NewTableOptions(40, 30, 0.5, 0.5, result.FontSize, 0.5, fontPath, result.Language, result.Theme, "Asia/Shanghai", FontBytes)
	table, err := render.NewTableWithOption(result.Nodes, &options)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	linksCount := 0
	successCount := 0
	for _, v := range result.Nodes {
		linksCount += 1
		if v.IsOk {
			successCount += 1
		}
	}
	msg := table.FormatTraffic(result.TotalTraffic, result.TotalTime, fmt.Sprintf("%d/%d", successCount, linksCount))
	if picdata, err := table.EncodeB64(msg); err == nil {
		fmt.Fprint(w, picdata)
	}

}

func isPrivateIP(ip net.IP) bool {
	var privateIPBlocks []*net.IPNet
	for _, cidr := range []string{
		// don't check loopback ips
		//"127.0.0.0/8",    // IPv4 loopback
		//"::1/128",        // IPv6 loopback
		//"fe80::/10",      // IPv6 link-local
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

func localIP() (string, error) {
	addresses, err := listAddresses()
	if err != nil {
		return "", err
	}
	if len(addresses) < 1 {
		return "", errors.New("IP not found")
	}

	return addresses[0], nil
}

func listAddresses() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		ifAddrs, _ := iface.Addrs()
		for _, ifAddr := range ifAddrs {
			switch v := ifAddr.(type) {
			case *net.IPNet:
				addresses = append(addresses, v.IP.String())
			case *net.IPAddr:
				addresses = append(addresses, v.IP.String())
			}
		}
	}

	return addresses, nil
}

type GetSubscriptionLink struct {
	FilePath string `json:"filePath"`
	Group    string `json:"group"`
}

var subscriptionLinkMap map[string]string = make(map[string]string)

func getSubscriptionLink(w http.ResponseWriter, r *http.Request) {
	body := GetSubscriptionLink{}
	if r.Body == nil {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	if err = json.Unmarshal(data, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(body.FilePath) == 0 || len(body.Group) == 0 {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	ipAddr, err := localIP()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(body.FilePath)))
	subscriptionLinkMap[md5Hash] = body.FilePath
	subscriptionLink := fmt.Sprintf("http://%s:10888/getSubscription?key=%s&group=%s", ipAddr, md5Hash, body.Group)
	fmt.Fprint(w, subscriptionLink)
}

func getSubscription(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	key := queries.Get("key")
	if len(key) < 1 {
		http.Error(w, "Key not found", 400)
		return
	}
	// sub format
	sub := queries.Get("sub")
	filePath, ok := subscriptionLinkMap[key]
	if !ok {
		http.Error(w, "Wrong key", 400)
		return
	}
	// convert yaml link
	if isYamlFile(filePath) && utils.IsUrl(filePath) {
		links, err := getSubscriptionLinks(filePath)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		b64Data := base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
		w.Write([]byte(b64Data))
		return
	}
	// FIXME
	if isYamlFile(filePath) {
		data, err := writeClash(filePath)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write(data)
		return
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		// return raw data
		if len(filePath) > 4096 {
			w.Write([]byte(filePath))
			return
		}
		http.Error(w, err.Error(), 400)
		return
	}

	if len(data) > 128 && strings.Contains(string(data[:128]), "proxies:") {
		if dataClash, err := writeClash(filePath); err == nil && len(dataClash) > 0 {
			data = dataClash
		}
	}
	// convert shadowrocket to v2ray
	if sub == "v2ray" {
		if dataShadowrocket, err := writeShadowrocket(data); err == nil && len(dataShadowrocket) > 0 {
			data = dataShadowrocket
		}
	}

	w.Write(data)
}

func getUserConfig(w http.ResponseWriter, r *http.Request) {
	// p, err := getUserConfigFilepath(base64.RawStdEncoding.EncodeToString([]byte(r.RemoteAddr)))
	p, err := getUserConfigFilepath(strings.Split(r.RemoteAddr, ":")[0])
	if err != nil {
		http.Error(w, "fail to get user config file path", 400)
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		http.Error(w, "fail to read user config file ", 400)
		return
	}
	w.Write(data)
}

func getUserConfigFilepath(suffix string) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(ex), fmt.Sprintf("userConfig%s.json", suffix)), nil
}

func saveUserOptions(options *ProfileTestOptions, suffix string) {
	p, err := getUserConfigFilepath(suffix)
	if err != nil {
		return
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	defer f.Close()
	if data, err := json.Marshal(options); err == nil {
		f.Write(data)
	}
}

type StartProxyBody struct {
	Links []string `json:"links"`
	Port  uint16   `json:"port"`
}

func startProxy(w http.ResponseWriter, r *http.Request) {
	body := StartProxyBody{}
	if r.Body == nil {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	if err = json.Unmarshal(data, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(body.Links) == 0 || body.Port < 80 || body.Port > 65535 {
		http.Error(w, "Invalid Parameter", 400)
		return
	}
	// close previous proxy
	if prevProxy != nil {
		prevProxy.Close()
		prevProxy = nil
	}
	// start proxy
	for _, link := range body.Links {
		ch := make(chan int64, 1)
		c := core.Config{
			LocalHost: "0.0.0.0",
			LocalPort: int(body.Port),
			Link:      link,
			Ping:      2,
			PingCh:    ch,
		}
		p, err := core.StartInstance(c)
		if err != nil {
			continue
		}
		elapse := <-ch
		if elapse == 0 {
			if p != nil {
				p.Close()
			}
			continue
		}
		go p.Run()
		prevProxy = p
		w.Write([]byte("Start Proxy"))
		return
	}
	http.Error(w, "Start Proxy Failed", 400)
}

func writeClash(filePath string) ([]byte, error) {
	links, err := parseClashFileByLine(filePath)
	if err != nil {
		//
		return nil, err
	}
	subscription := []byte(strings.Join(links, "\n"))
	data := make([]byte, base64.StdEncoding.EncodedLen(len(subscription)))
	base64.StdEncoding.Encode(data, subscription)
	return data, nil
}

func writeShadowrocket(data []byte) ([]byte, error) {
	links, err := ParseLinks(string(data))
	if err != nil {
		return nil, err
	}
	newLinks := make([]string, 0, len(links))
	for _, link := range links {
		if strings.HasPrefix(link, "vmess://") && strings.Contains(link, "&") {
			if newLink, err := config.ShadowrocketLinkToVmessLink(link); err == nil {
				newLinks = append(newLinks, newLink)
			}
		} else {
			newLinks = append(newLinks, link)
		}
	}
	subscription := []byte(strings.Join(newLinks, "\n"))
	data = make([]byte, base64.StdEncoding.EncodedLen(len(subscription)))
	base64.StdEncoding.Encode(data, subscription)
	return data, nil
}
