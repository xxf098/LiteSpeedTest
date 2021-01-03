package main

import (
	"context"
	"errors"
	"strings"

	"github.com/xxf098/lite-proxy/component/resolver"
	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/dns"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/adapter"
	"github.com/xxf098/lite-proxy/tunnel/http"
	"github.com/xxf098/lite-proxy/tunnel/socks"
)

func startInstance(c Config) (*proxy.Proxy, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "LocalHost", c.LocalHost)
	ctx = context.WithValue(ctx, "LocalPort", c.LocalPort)
	adapterServer, err := adapter.NewServer(ctx, nil)
	if err != nil {
		return nil, err
	}
	httpServer, err := http.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}
	socksServer, err := socks.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}
	sources := []tunnel.Server{httpServer, socksServer}
	sink, err := createSink(ctx, c.Link)
	if err != nil {
		return nil, err
	}
	setDefaultResolver()
	p := proxy.NewProxy(ctx, cancel, sources, sink)
	return p, nil
}

func createSink(ctx context.Context, link string) (tunnel.Client, error) {
	var d proxy.ContextDialer
	if strings.HasPrefix(link, "vmess://") {
		vmessOption, err := config.VmessLinkToVmessOption(link)
		if err != nil {
			return nil, err
		}
		d, err = outbound.NewVmess(*vmessOption)
		if err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(link, "trojan://") {
		trojanOption, err := config.TrojanLinkToTrojanOption(link)
		if err != nil {
			return nil, err
		}
		d, err = outbound.NewTrojan(trojanOption)
		if err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(link, "ss://") {
		ssOption, err := config.SSLinkToSSOption(link)
		if err != nil {
			return nil, err
		}
		d, err = outbound.NewShadowSocks(ssOption)
		if err != nil {
			return nil, err
		}
	}
	if d != nil {
		return proxy.NewClient(ctx, d), nil
	}

	return nil, errors.New("not supported link")
}

func setDefaultResolver() {
	resolver.DefaultResolver = dns.DefaultResolver()
}
