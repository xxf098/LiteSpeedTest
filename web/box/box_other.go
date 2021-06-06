package box

import (
	"fmt"

	"github.com/gobuffalo/packr/v2"
)

var (
	gui  = packr.New("gui", "../gui")
	misc = packr.New("misc", "../misc")
	wasm = packr.New("wasm", "../wasm")
)

func GetBox(name string) (*packr.Box, error) {
	switch name {
	case "gui":
		return gui, nil
	case "misc":
		return misc, nil
	case "wasm":
		return wasm, nil
	default:
		return nil, fmt.Errorf("box not found")
	}
}
