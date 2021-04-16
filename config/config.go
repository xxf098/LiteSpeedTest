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
		option, err1 := VmessLinkToVmessOption(link)
		if err1 != nil {
			return nil, err1
		}
		d, err = outbound.NewVmess(option)
	case "trojan":
		option, err1 := TrojanLinkToTrojanOption(link)
		if err1 != nil {
			return nil, err1
		}
		d, err = outbound.NewTrojan(option)
	case "ss":
		option, err1 := SSLinkToSSOption(link)
		if err1 != nil {
			return nil, err1
		}
		d, err = outbound.NewShadowSocks(option)
	case "ssr":
		option, err1 := SSRLinkToSSROption(link)
		if err1 != nil {
			return nil, err1
		}
		d, err = outbound.NewShadowSocksR(option)
	default:
		return nil, common.NewError("Not Suported Link")
	}
	return d, err
}

type Config struct {
	Protocol string
	Remarks  string
	Server   string
	Port     int
}

func Link2Config(link string) (*Config, error) {
	var cfg *Config = &Config{
		Protocol: "unknown",
		Remarks:  "",
		Server:   "127.0.0.1",
		Port:     80,
	}

	cfgVmess, err := VmessLinkToVmessConfigIP(link, false)
	if err == nil {
		cfg = &Config{
			Protocol: "vmess",
			Remarks:  cfgVmess.Ps,
			Server:   cfgVmess.Add,
			Port:     cfgVmess.PortInt,
		}
		return cfg, nil
	}
	cfgSSR, err := SSRLinkToSSROption(link)
	if err == nil {
		cfg = &Config{
			Protocol: "ssr",
			Remarks:  cfgSSR.Remarks,
			Server:   cfgSSR.Server,
			Port:     cfgSSR.Port,
		}
		return cfg, nil
	}
	cfgTrojan, err := TrojanLinkToTrojanOption(link)
	if err == nil {
		cfg = &Config{
			Protocol: "trojan",
			Remarks:  cfgTrojan.Remarks,
			Server:   cfgTrojan.Server,
			Port:     cfgTrojan.Port,
		}
		return cfg, nil
	}
	cfgSS, err := SSLinkToSSOption(link)
	if err == nil {
		cfg.Protocol = "ss"
		cfg.Remarks = cfgSS.Remarks
		cfg = &Config{
			Protocol: "ss",
			Remarks:  cfgSS.Remarks,
			Server:   cfgSS.Server,
			Port:     cfgSS.Port,
		}
		return cfg, nil
	}
	return cfg, err
}
