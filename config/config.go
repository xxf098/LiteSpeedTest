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
	switch strings.ToLower(matches[1]) {
	case "vmess":
		option, err := VmessLinkToVmessOption(link)
		if err != nil {
			return nil, err
		}
		d, err := outbound.NewVmess(option)
		if err != nil {
			return nil, err
		}
		return d, nil
	case "trojan":
		option, err := TrojanLinkToTrojanOption(link)
		if err != nil {
			return nil, err
		}
		d, err := outbound.NewTrojan(option)
		if err != nil {
			return nil, err
		}
		return d, nil
	case "ss":
		option, err := SSLinkToSSOption(link)
		if err != nil {
			return nil, err
		}
		d, err := outbound.NewShadowSocks(option)
		if err != nil {
			return nil, err
		}
		return d, nil
	case "ssr":
		option, err := SSRLinkToSSROption(link)
		if err != nil {
			return nil, err
		}
		d, err := outbound.NewShadowSocksR(option)
		if err != nil {
			return nil, err
		}
		return d, nil
	default:
		return nil, common.NewError("Not Suported Link")
	}
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
