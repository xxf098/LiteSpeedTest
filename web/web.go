package web

import (
	"fmt"
	"net/http"
)

func ServeFile() {
	http.Handle("/", http.FileServer(http.Dir("web/gui/")))
	fmt.Println("Start server at http://127.0.0.1:10871")
	http.ListenAndServe(":10871", nil)
}
