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
