package config

import (
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/outbound"
)

func SSLinkToSSOption(link string) (*outbound.ShadowSocksOption, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "ss" {
		return nil, errors.New("not a shadowsocks link")
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
	data, err := base64.StdEncoding.DecodeString(pass)
	if err != nil {
		if data, err = base64.RawStdEncoding.DecodeString(pass); err != nil {
			return nil, err
		}
	}
	userinfo := string(data)
	splits := strings.SplitN(userinfo, ":", 2)
	method := splits[0]
	pass = splits[1]

	shadwosocksOption := &outbound.ShadowSocksOption{
		Name:     "ss",
		Server:   host,
		Port:     port,
		Password: pass,
		Cipher:   method,
	}
	return shadwosocksOption, nil
}

func init() {
	outbound.RegisterDialerCreator("ss", func(link string) (outbound.Dialer, error) {
		ssOption, err := SSLinkToSSOption(link)
		if err != nil {
			return nil, err
		}
		return outbound.NewShadowSocks(ssOption)
	})
}
