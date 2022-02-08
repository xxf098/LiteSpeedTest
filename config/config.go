package config

import (
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
			cfg = &Config{
				Protocol: "vmess",
				Remarks:  cfgVmess.Ps,
				Server:   cfgVmess.Add,
				Port:     cfgVmess.PortInt,
				Net:      cfgVmess.Net,
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
			}
		}
	default:
		return nil, common.NewError("Not Suported Link")
	}
	return cfg, err
}
