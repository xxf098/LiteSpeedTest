package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/common/structure"
	"github.com/xxf098/lite-proxy/outbound"
)

func ParseProxy(mapping map[string]interface{}, namePrefix string) (string, error) {
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
		link = fmt.Sprintf("ss://%s@%s", base64.StdEncoding.EncodeToString([]byte(auth)), net.JoinHostPort(ssOption.Server, strconv.Itoa(ssOption.Port)))
		if len(ssOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s", link, url.QueryEscape(ssOption.Name))
		}
	case "ssr":
		ssrOption := &outbound.ShadowSocksROption{}
		err = decoder.Decode(mapping, ssrOption)
		if err != nil {
			break
		}
		password := base64.StdEncoding.EncodeToString([]byte(ssrOption.Password))
		link = fmt.Sprintf("%s:%s:%s:%s:%s", net.JoinHostPort(ssrOption.Server, strconv.Itoa(ssrOption.Port)), ssrOption.Protocol, ssrOption.Cipher, ssrOption.Obfs, password)
		remarks := base64.StdEncoding.EncodeToString([]byte(ssrOption.Name))

		obfsParam := base64.StdEncoding.EncodeToString([]byte(ssrOption.ObfsParam))
		protocolParam := base64.StdEncoding.EncodeToString([]byte(ssrOption.ProtocolParam))
		link = fmt.Sprintf("%s/?obfsparam=%s&remarks=%s&protoparam=%s", link, url.QueryEscape(obfsParam), url.QueryEscape(remarks), url.QueryEscape(protocolParam))
		link = fmt.Sprintf("ssr://%s", base64.StdEncoding.EncodeToString([]byte(link)))
	case "vmess":
		// TODO: h2
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
			tls = "tls"
		}
		if len(vmessOption.WSPath) == 0 && len(vmessOption.WSOpts.Path) > 0 {
			vmessOption.WSPath = vmessOption.WSOpts.Path
		}
		host := ""
		if h, ok := vmessOption.WSHeaders["Host"]; ok {
			host = h
		} else {
			headers := vmessOption.WSOpts.Headers
			if h, ok := headers["Host"]; ok {
				host = h
			}
		}
		if len(vmessOption.Network) < 1 {
			vmessOption.Network = "tcp"
		}
		if len(vmessOption.Cipher) < 1 {
			vmessOption.Cipher = "none"
		}
		skipCertVerify := vmessOption.SkipCertVerify
		if len(vmessOption.ServerName) < 1 {
			skipCertVerify = true
		}
		id := vmessOption.UUID
		if len(id) < 1 {
			id = vmessOption.Password
		}
		c := VmessConfigMarshal{
			Ps:             namePrefix + vmessOption.Name,
			Add:            vmessOption.Server,
			Port:           vmessOption.Port,
			Aid:            vmessOption.AlterID,
			ID:             id,
			Type:           vmessOption.Cipher,
			TLS:            tls,
			Net:            vmessOption.Network,
			Path:           vmessOption.WSPath,
			Host:           host,
			SkipCertVerify: skipCertVerify,
			ServerName:     vmessOption.ServerName,
			Security:       vmessOption.Cipher,
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

		link = fmt.Sprintf("trojan://%s@%s", trojanOption.Password, net.JoinHostPort(trojanOption.Server, strconv.Itoa(trojanOption.Port)))
		query := []string{}
		// allowInsecure
		if trojanOption.SkipCertVerify {
			query = append(query, "allowInsecure=1")
		} else {
			query = append(query, "security=tls")
		}
		if len(trojanOption.SNI) > 0 {
			query = append(query, fmt.Sprintf("sni=%s", trojanOption.SNI))
		}
		// ws query
		if trojanOption.Network == "ws" {
			query = append(query, "type=ws")
			if len(trojanOption.WSOpts.Path) > 0 {
				query = append(query, fmt.Sprintf("path=%s", trojanOption.WSOpts.Path))
				for k, v := range trojanOption.WSOpts.Headers {
					query = append(query, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
		// grpc
		if trojanOption.Network == "grpc" {
			query = append(query, "type=grpc")
			if len(trojanOption.GrpcOpts.GrpcServiceName) > 0 {
				query = append(query, fmt.Sprintf("serviceName=%s", trojanOption.GrpcOpts.GrpcServiceName))
			}
		}

		if len(query) > 0 {
			link = fmt.Sprintf("%s?%s", link, strings.Join(query, "&"))
		}
		if len(trojanOption.Remarks) > 0 {
			link = fmt.Sprintf("%s#%s%s", link, namePrefix, url.QueryEscape(trojanOption.Remarks))
		}
		if len(trojanOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s%s", link, namePrefix, url.QueryEscape(trojanOption.Name))
		}
	case "http":
		httpOption := &outbound.HttpOption{}
		err = decoder.Decode(mapping, httpOption)
		if err != nil {
			break
		}
		link = fmt.Sprintf("http://%s@%s", httpOption.Password, net.JoinHostPort(httpOption.Server, strconv.Itoa(httpOption.Port)))
		query := []string{}
		query = append(query, fmt.Sprintf("tls=%t", httpOption.TLS))
		if len(httpOption.UserName) > 0 {
			query = append(query, fmt.Sprintf("username=%s", httpOption.UserName))
		}
		// allowInsecure
		if httpOption.SkipCertVerify {
			query = append(query, "allowInsecure=1")
		}
		if len(httpOption.SNI) > 0 {
			query = append(query, fmt.Sprintf("sni=%s", httpOption.SNI))
		}
		if len(query) > 0 {
			link = fmt.Sprintf("%s?%s", link, strings.Join(query, "&"))
		}
		if len(httpOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s%s", link, namePrefix, url.QueryEscape(httpOption.Name))
		}
	default:
		return "", fmt.Errorf("unsupport proxy type: %s", proxyType)
	}

	if err != nil {
		return "", err
	}

	return link, nil

}
