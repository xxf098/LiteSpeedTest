package web

import "embed"

//go:embed index.html style.css favicon.ico main.wasm wasm_exec.js
var guiStatic embed.FS

//go:embed misc/WenQuanYiMicroHei-01.ttf
var FontBytes []byte
