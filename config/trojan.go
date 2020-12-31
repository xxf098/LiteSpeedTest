package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

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
	splits := strings.SplitN(hostport, ":", 2)
	host := splits[0]
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
		ALPN:     []string{"h2", "http/1.1"},
	}
	if rawQuery, err := url.ParseQuery(u.RawQuery); err == nil {
		trojanOption.SNI = rawQuery.Get("sni")
	}
	return &trojanOption, nil
}
