package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/skip2/go-qrcode"
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
	items := []Item{}
	jsonItems := inputs[0].String()
	log.Println(jsonItems)
	err := json.Unmarshal([]byte(jsonItems), &items)
	if err != nil {
		return nil
	}
	ch := make(chan struct{}, 3)
	for _, v := range items {
		go func(eid string, text string, size int) {
			defer func() {
				ch <- struct{}{}
			}()
			document := js.Global().Get("document")
			elem := document.Call("getElementById", eid)
			elem.Call("setAttribute", "title", text)
			var bytes []byte
			bytes, err := qrcode.Encode(text, qrcode.High, size)
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

func main() {
	js.Global().Set("printMessage", js.FuncOf(printMessage))
	js.Global().Set("wasmQRcode", js.FuncOf(wasmQRcode))
	select {}
}
