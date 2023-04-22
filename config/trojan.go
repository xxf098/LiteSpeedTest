package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"

	"github.com/xxf098/lite-proxy/outbound"
)

// type TrojanGoOption struct {
// }

// TODO: SkipCertVerify
func TrojanLinkToTrojanOption(link string) (*outbound.TrojanOption, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "trojan" {
		return nil, errors.New("not a trojan link")
	}
	pass := u.User.Username()
	hostport := u.Host
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	// frag := u.Fragment
	// fmt.Printf("password: %s, host: %s, port: %d, frag: %s\n", pass, host, port, frag)
	trojanOption := outbound.TrojanOption{
		Name:     "trojan",
		Password: pass,
		Port:     port,
		Server:   host,
		Remarks:  u.Fragment,
		ALPN:     []string{"h2", "http/1.1"},
	}
	if rawQuery, err := url.ParseQuery(u.RawQuery); err == nil {
		trojanOption.SNI = rawQuery.Get("sni")
		trojanOption.SkipCertVerify = rawQuery.Get("allowInsecure") == "1"
		network := rawQuery.Get("type")
		if network == "ws" {
			trojanOption.Network = "ws"
			p := rawQuery.Get("path")
			if len(p) > 0 {
				wsOption := outbound.WSOptions{Path: p}
				wsOption.Headers = map[string]string{}
				host := rawQuery.Get("host")
				if len(host) > 0 {
					wsOption.Headers["host"] = host
				}
				Host := rawQuery.Get("Host")
				if len(Host) > 0 {
					wsOption.Headers["Host"] = Host
				}
				trojanOption.WSOpts = wsOption
			}
		}
		if network == "grpc" {
			trojanOption.Network = "grpc"
			serviceName := rawQuery.Get("serviceName")
			if len(serviceName) > 0 {
				trojanOption.GrpcOpts = outbound.GrpcOptions{GrpcServiceName: serviceName}
			}
		}

	}
	if len(trojanOption.Remarks) < 1 {
		trojanOption.Remarks = trojanOption.Server
	}
	return &trojanOption, nil
}

func init() {
	outbound.RegisterDialerCreator("trojan", func(link string) (outbound.Dialer, error) {
		trojanOption, err := TrojanLinkToTrojanOption(link)
		if err != nil {
			return nil, err
		}
		return outbound.NewTrojan(trojanOption)
	})
}
