package config

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/utils"
)

func Link2Dialer(link string) (outbound.Dialer, error) {
	matches, err := utils.CheckLink(link)
	if err != nil {
		return nil, err
	}
	var d outbound.Dialer
	switch strings.ToLower(matches[1]) {
	case "vmess":
		if option, err1 := VmessLinkToVmessOption(link); err1 != nil {
			return nil, err1
		} else {
			d, err = outbound.NewVmess(option)
		}
	case "trojan":
		if option, err1 := TrojanLinkToTrojanOption(link); err1 != nil {
			return nil, err1
		} else {
			d, err = outbound.NewTrojan(option)
		}
	case "ss":
		if option, err1 := SSLinkToSSOption(link); err1 != nil {
			return nil, err1
		} else {
			d, err = outbound.NewShadowSocks(option)
		}
	case "ssr":
		if option, err1 := SSRLinkToSSROption(link); err1 != nil {
			return nil, err1
		} else {
			d, err = outbound.NewShadowSocksR(option)
		}
	default:
		return nil, common.NewError("Not Suported Link")
	}
	return d, err
}

type Config struct {
	Protocol string
	Remarks  string
	Server   string
	Net      string // vmess net type
	Port     int
	Password string
	SNI      string
}

func Link2Config(link string) (*Config, error) {
	matches, err := utils.CheckLink(link)
	if err != nil {
		return nil, err
	}
	var cfg *Config = &Config{
		Protocol: "unknown",
		Remarks:  "",
		Server:   "127.0.0.1",
		Port:     80,
	}
	switch strings.ToLower(matches[1]) {
	case "vmess":
		if cfgVmess, err1 := VmessLinkToVmessConfigIP(link, false); err1 != nil {
			return nil, err1
		} else {
			remarks := cfgVmess.Ps
			if len(remarks) < 1 {
				remarks = cfgVmess.Add
			}
			cfg = &Config{
				Protocol: "vmess",
				Remarks:  remarks,
				Server:   cfgVmess.Add,
				Port:     cfgVmess.PortInt,
				Net:      cfgVmess.Net,
				Password: cfgVmess.ID,
				SNI:      cfgVmess.Host,
			}
		}
	case "trojan":
		if cfgTrojan, err1 := TrojanLinkToTrojanOption(link); err1 != nil {
			return nil, err1
		} else {
			cfg = &Config{
				Protocol: "trojan",
				Remarks:  cfgTrojan.Remarks,
				Server:   cfgTrojan.Server,
				Port:     cfgTrojan.Port,
				Password: cfgTrojan.Password,
				Net:      cfgTrojan.Network,
				SNI:      cfgTrojan.SNI,
			}
		}
	case "http":
		if cfgHttp, err1 := HttpLinkToHttpOption(link); err1 != nil {
			return nil, err1
		} else {
			cfg = &Config{
				Protocol: "http",
				Remarks:  cfgHttp.Remarks,
				Server:   cfgHttp.Server,
				Port:     cfgHttp.Port,
				Password: cfgHttp.Password,
			}
		}
	case "ss":
		if cfgSS, err1 := SSLinkToSSOption(link); err1 != nil {
			return nil, err1
		} else {
			cfg = &Config{
				Protocol: "ss",
				Remarks:  cfgSS.Remarks,
				Server:   cfgSS.Server,
				Port:     cfgSS.Port,
				Password: cfgSS.Password,
			}
		}
	case "ssr":
		if cfgSSR, err1 := SSRLinkToSSROption(link); err1 != nil {
			return nil, err1
		} else {
			cfg = &Config{
				Protocol: "ssr",
				Remarks:  cfgSSR.Remarks,
				Server:   cfgSSR.Server,
				Port:     cfgSSR.Port,
				Password: cfgSSR.Password,
			}
		}
	case "vless":
		u, err := url.Parse(link)
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, err
		}
		password, _ := u.User.Password()
		cfg = &Config{
			Protocol: "vless",
			Remarks:  u.Fragment,
			Server:   u.Host,
			Port:     port,
			Password: password,
		}
	default:
		return nil, common.NewError("Not Suported Link")
	}
	return cfg, err
}
