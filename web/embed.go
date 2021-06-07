package web

import "embed"

//go:embed index.html style.css favicon.ico
var guiStatic embed.FS

//go:embed wasm/*
var wasmStatic embed.FS

//go:embed misc/WenQuanYiMicroHei-01.ttf
var fontBytes []byte
