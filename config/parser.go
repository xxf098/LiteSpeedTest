package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/xxf098/lite-proxy/common/structure"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/utils"
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
		tls := ""
		if vmessOption.TLS {
			tls = ""
		}
		host := ""
		if h, ok := vmessOption.WSHeaders["Host"]; ok {
			host = h
		}
		c := VmessConfig{
			Ps:   vmessOption.Name,
			Add:  vmessOption.Server,
			Port: []byte(utils.U16toa(vmessOption.Port)),
			Aid:  []byte(strconv.Itoa(vmessOption.AlterID)),
			ID:   vmessOption.UUID,
			Type: vmessOption.Cipher,
			TLS:  tls,
			Net:  vmessOption.Network,
			Path: vmessOption.WSPath,
			Host: host,
		}
		data, err := json.MarshalIndent(&c, "", "    ")
		if err != nil {
			return "", err
		}
		link = fmt.Sprintf("vmess://%s", base64.StdEncoding.EncodeToString(data))
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
