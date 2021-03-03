package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func ServeFile() {
	http.Handle("/", http.FileServer(http.Dir("web/gui/")))
	http.HandleFunc("/test", updateTest)
	fmt.Println("Start server at http://127.0.0.1:10871")
	http.ListenAndServe(":10871", nil)
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
		splits := strings.SplitN(string(message), "^", 2)
		if len(splits) < 2 {
			break
		}
		links, err := parseLinks(splits[0])
		if err != nil {
			break
		}
		options, err := parseOptions(splits[1])
		if err != nil {
			break
		}
		p := ProfileTest{
			Conn:        c,
			MessageType: mt,
			Links:       links,
			Options:     options,
		}
		go p.testAll(ctx)
		// err = c.WriteMessage(mt, getMsgByte(0, "gotspeed"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
