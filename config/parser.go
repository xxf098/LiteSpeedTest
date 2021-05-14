package config

import (
	"fmt"

	"github.com/xxf098/lite-proxy/common/structure"
	"github.com/xxf098/lite-proxy/outbound"
)

func ParseProxy(mapping map[string]interface{}) (string, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "proxy", WeaklyTypedInput: true})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return "", fmt.Errorf("missing type")
	}

	var (
		err  error
		link string
	)

	switch proxyType {
	case "ss":
		ssOption := &outbound.ShadowSocksOption{}
		err = decoder.Decode(mapping, ssOption)
		if err != nil {
			break
		}

	case "ssr":
		ssrOption := &outbound.ShadowSocksROption{}
		err = decoder.Decode(mapping, ssrOption)
		if err != nil {
			break
		}
	case "vmess":
		vmessOption := &outbound.VmessOption{
			HTTPOpts: outbound.HTTPOptions{
				Method: "GET",
				Path:   []string{"/"},
			},
		}
		err = decoder.Decode(mapping, vmessOption)
		if err != nil {
			break
		}
	case "trojan":
		trojanOption := &outbound.TrojanOption{}
		err = decoder.Decode(mapping, trojanOption)
		if err != nil {
			break
		}

	default:
		return "", fmt.Errorf("unsupport proxy type: %s", proxyType)
	}

	if err != nil {
		return "", err
	}

	return link, nil

}
