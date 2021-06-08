package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"syscall/js"

	qrcode "github.com/skip2/go-qrcode"
)

func printMessage(this js.Value, inputs []js.Value) interface{} {
	callback := inputs[len(inputs)-1]
	message := inputs[0].String()
	callback.Invoke(js.Null(), strings.ToUpper(message))
	return nil
}

func wasmQRcode(this js.Value, inputs []js.Value) interface{} {
	eid := inputs[0].String()
	text := inputs[1].String()
	size := inputs[2].Int()
	document := js.Global().Get("document")
	elem := document.Call("getElementById", eid)
	elem.Call("setAttribute", "title", text)
	var bytes []byte
	bytes, err := qrcode.Encode(text, qrcode.Medium, size)
	if err != nil {
		return nil
	}
	imgData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes)
	html := fmt.Sprintf(`<canvas width="%d" height="%d" style="display: none;"></canvas>
	<img style="display: block;" src="%s">`, size, size, imgData)
	elem.Set("innerHTML", html)
	return nil
}

func main() {
	js.Global().Set("printMessage", js.FuncOf(printMessage))
	js.Global().Set("wasmQRcode", js.FuncOf(wasmQRcode))
	<-make(chan bool)
}
