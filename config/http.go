package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"

	"github.com/xxf098/lite-proxy/outbound"
)

func HttpLinkToHttpOption(link string) (*outbound.HttpOption, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" {
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
	httpOption := outbound.HttpOption{
		Name:     "http",
		Remarks:  u.Fragment,
		Password: pass,
		Port:     port,
		Server:   host,
	}
	if rawQuery, err := url.ParseQuery(u.RawQuery); err == nil {
		httpOption.UserName = rawQuery.Get("username")
		httpOption.TLS = rawQuery.Get("tls") == "true"
		httpOption.SNI = rawQuery.Get("sni")
		httpOption.SkipCertVerify = rawQuery.Get("allowInsecure") == "1"
	}
	return &httpOption, nil
}

func init() {
	outbound.RegisterDialerCreator("http", func(link string) (outbound.Dialer, error) {
		httpOption, err := HttpLinkToHttpOption(link)
		if err != nil {
			return nil, err
		}
		return outbound.NewHttp(*httpOption), nil
	})
}
