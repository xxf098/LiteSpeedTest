package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
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
		auth := fmt.Sprintf("%s:%s", ssOption.Cipher, ssOption.Password)
		link = fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(auth)), ssOption.Server, ssOption.Port)
		if len(ssOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s", link, url.QueryEscape(ssOption.Name))
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
		if len(vmessOption.Network) < 1 {
			vmessOption.Network = "tcp"
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
		// TODO: SNI
		link = fmt.Sprintf("trojan://%s@%s:%d", trojanOption.Password, trojanOption.Server, trojanOption.Port)
		if len(trojanOption.Remarks) > 0 {
			link = fmt.Sprintf("%s#%s", link, url.QueryEscape(trojanOption.Remarks))
		}
		if len(trojanOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s", link, url.QueryEscape(trojanOption.Name))
		}
	default:
		return "", fmt.Errorf("unsupport proxy type: %s", proxyType)
	}

	if err != nil {
		return "", err
	}

	return link, nil

}
