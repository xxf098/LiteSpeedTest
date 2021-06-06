package main

import (
	"strings"
	"syscall/js"
)

func printMessage(this js.Value, inputs []js.Value) interface{} {
	callback := inputs[len(inputs)-1]
	message := inputs[0].String()
	callback.Invoke(js.Null(), strings.ToUpper(message))
	return nil
}

func main() {
	js.Global().Set("printMessage", js.FuncOf(printMessage))
	<-make(chan bool)
}
