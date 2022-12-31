//go:build !gomobile
// +build !gomobile

package web

import "embed"

//go:embed gui/dist/*
var guiStatic embed.FS

//go:embed misc/WenQuanYiMicroHei-01.ttf
var FontBytes []byte
