package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"syscall/js"

	"github.com/skip2/go-qrcode"
	"github.com/xxf098/lite-proxy/web"
	"github.com/xxf098/lite-proxy/web/render"
)

func printMessage(this js.Value, inputs []js.Value) interface{} {
	callback := inputs[len(inputs)-1]
	message := inputs[0].String()
	callback.Invoke(js.Null(), strings.ToUpper(message))
	return nil
}

type Item struct {
	Gid  string `json:"gid"`
	Link string `json:"link"`
	Size int    `json:"size"`
}

func wasmQRcode(this js.Value, inputs []js.Value) interface{} {
	// eid := inputs[0].String()
	// text := inputs[1].String()
	// size := inputs[2].Int()
	items := []Item{}
	jsonItems := inputs[0].String()
	log.Println(jsonItems)
	err := json.Unmarshal([]byte(jsonItems), &items)
	if err != nil {
		return nil
	}
	ch := make(chan bool, 2)
	for _, v := range items {
		go func(eid string, text string, size int) {
			defer func() {
				ch <- true
			}()
			document := js.Global().Get("document")
			elem := document.Call("getElementById", eid)
			elem.Call("setAttribute", "title", text)
			var bytes []byte
			bytes, err := qrcode.Encode(text, qrcode.Medium, size)
			if err != nil {
				return
			}
			imgData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes)
			html := fmt.Sprintf(`<canvas width="%d" height="%d" style="display: none;"></canvas>
				<img style="display: block;" src="%s">`, size, size, imgData)
			elem.Set("innerHTML", html)
		}(v.Gid, v.Link, v.Size)
		<-ch
	}

	return nil
}

func wasmResultTable(this js.Value, inputs []js.Value) interface{} {
	nodes := make([]render.Node, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = render.Node{
			Group:    "节点列表",
			Remarks:  fmt.Sprintf("美国加利福尼亚免费测试%d", i),
			Protocol: "vmess",
			Ping:     fmt.Sprintf("%d", rand.Intn(800-50)+50),
			AvgSpeed: int64((rand.Intn(20-1) + 1) * 1024 * 1024),
			MaxSpeed: int64((rand.Intn(60-5) + 5) * 1024 * 1024),
		}
	}
	options := render.NewTableOptions(40, 30, 0.5, 0.5, 28, 0.5, "WenQuanYiMicroHei-01.ttf", "en", "rainbow", "Asia/Shanghai", web.FontBytes)
	table, err := render.NewTableWithOption(nodes, &options)
	if err != nil {
		return nil
	}
	msg := table.FormatTraffic("10.2G", "3m13s", "50/50")
	encodePNG, err := table.EncodeB64(msg)
	if err != nil {
		return nil
	}
	// document := js.Global().Get("document")
	// elem := document.Call("getElementById", "result_png")
	// html := fmt.Sprintf(`<img class="el-image__inner" src="%s">`, encodePNG)
	// elem.Call("setAttribute", "src", encodePNG)
	// elem.Set("innerHTML", html)
	return encodePNG
}

func main() {
	js.Global().Set("printMessage", js.FuncOf(printMessage))
	js.Global().Set("wasmQRcode", js.FuncOf(wasmQRcode))
	js.Global().Set("wasmResultTable", js.FuncOf(wasmResultTable))
	<-make(chan bool)
}
