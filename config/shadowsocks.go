package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/utils"
)

var (
	NotSSLink error = errors.New("not a shadowsocksR link")
)

func decodeB64SS(link string) (string, error) {
	if strings.Contains(link, "@") {
		return link, nil
	}
	regex := regexp.MustCompile(`^ss://([A-Za-z0-9+-=/_]+)`)
	res := regex.FindAllStringSubmatch(link, 1)
	b64 := ""
	if len(res) > 0 && len(res[0]) > 1 {
		b64 = res[0][1]
	}
	if b64 == "" {
		return link, nil
	}
	uri, err := utils.DecodeB64(b64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ss://%s", uri), nil
}

func SSLinkToSSOption(link1 string) (*outbound.ShadowSocksOption, error) {
	link, err := decodeB64SS(link1)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "ss" {
		return nil, errors.New("not a shadowsocks link")
	}
	pass := u.User.Username()
	hostport := u.Host
	host, port1, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(port1)
	if err != nil {
		return nil, err
	}
	userinfo, err := utils.DecodeB64(pass)
	if err != nil || !strings.Contains(userinfo, ":") {
		pw, _ := u.User.Password()
		if pw == "" {
			if err == nil {
				err = NotSSLink
			}
			return nil, err
		}
		userinfo = fmt.Sprintf("%s:%s", u.User.Username(), pw)
	}
	splits := strings.SplitN(userinfo, ":", 2)
	method := splits[0]
	pass = splits[1]
	remarks := u.Fragment
	if remarks == "" {
		// fmt.Println(link1)
		if splits := strings.Split(link1, "#"); len(splits) > 1 {
			if rmk, err := url.QueryUnescape(splits[1]); err == nil {
				remarks = rmk
			}
		}
	}

	shadwosocksOption := &outbound.ShadowSocksOption{
		Name:     "ss",
		Server:   host,
		Port:     port,
		Password: pass,
		Cipher:   method,
		Remarks:  remarks,
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
