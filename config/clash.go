package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/xxf098/lite-proxy/dns"
	"gopkg.in/yaml.v3"
)

// General config
type General struct {
	Inbound
	Controller
	Mode      string `json:"mode"`
	LogLevel  string `json:"log-level"`
	IPv6      bool   `json:"ipv6"`
	Interface string `json:"-"`
}

// Inbound
type Inbound struct {
	Port           int      `json:"port"`
	SocksPort      int      `json:"socks-port"`
	RedirPort      int      `json:"redir-port"`
	TProxyPort     int      `json:"tproxy-port"`
	MixedPort      int      `json:"mixed-port"`
	Authentication []string `json:"authentication"`
	AllowLan       bool     `json:"allow-lan"`
	BindAddress    string   `json:"bind-address"`
}

// Controller
type Controller struct {
	ExternalController string `json:"-"`
	ExternalUI         string `json:"-"`
	Secret             string `json:"-"`
}

// FallbackFilter config
type FallbackFilter struct {
	GeoIP  bool         `yaml:"geoip"`
	IPCIDR []*net.IPNet `yaml:"ipcidr"`
	Domain []string     `yaml:"domain"`
}

// Profile config
type Profile struct {
	StoreSelected bool `yaml:"store-selected"`
}

// Experimental config
type Experimental struct{}

// Config is clash config manager
type ClashConfig struct {
	General      *General
	Experimental *Experimental
	Profile      *Profile
	Proxies      []string
}

type RawFallbackFilter struct {
	GeoIP  bool     `yaml:"geoip"`
	IPCIDR []string `yaml:"ipcidr"`
	Domain []string `yaml:"domain"`
}

type ClashRawConfig struct {
	Port               int      `yaml:"port"`
	SocksPort          int      `yaml:"socks-port"`
	RedirPort          int      `yaml:"redir-port"`
	TProxyPort         int      `yaml:"tproxy-port"`
	MixedPort          int      `yaml:"mixed-port"`
	Authentication     []string `yaml:"authentication"`
	AllowLan           bool     `yaml:"allow-lan"`
	BindAddress        string   `yaml:"bind-address"`
	Mode               string   `yaml:"mode"`
	LogLevel           string   `yaml:"log-level"`
	NamePrefix         string   `yaml:"name-prefix"`
	IPv6               bool     `yaml:"ipv6"`
	ExternalController string   `yaml:"external-controller"`
	ExternalUI         string   `yaml:"external-ui"`
	Secret             string   `yaml:"secret"`
	Interface          string   `yaml:"interface-name"`

	ProxyProvider map[string]map[string]interface{} `yaml:"proxy-providers"`
	Hosts         map[string]string                 `yaml:"hosts"`
	Experimental  Experimental                      `yaml:"experimental"`
	Profile       Profile                           `yaml:"profile"`
	Proxy         []map[string]interface{}          `yaml:"proxies"`
	// ProxyGroup    []map[string]interface{}          `yaml:"proxy-groups"`
	// Rule          []string                          `yaml:"rules"`
}

type BaseProxy struct {
	Name   string      `yaml:"name"`
	Server string      `yaml:"server"`
	Port   interface{} `yaml:"port"` // int, string
	Type   string      `yaml:"type"`
}

func ParseBaseProxy(profile string) (*BaseProxy, error) {
	idx := strings.IndexByte(profile, byte('{'))
	idxLast := strings.LastIndexByte(profile, byte('}'))
	if idx < 0 || idxLast <= idx {
		// multiple lines form
		return nil, nil
	}
	p := profile[idx:]
	bp := &BaseProxy{}
	err := yaml.Unmarshal([]byte(p), bp)
	if err != nil {
		return nil, err
	}
	return bp, nil
}

// Parse config
func ParseClash(buf []byte) (*ClashConfig, error) {
	rawCfg, err := UnmarshalRawConfig(buf)
	if err != nil {
		return nil, err
	}

	return ParseRawConfig(rawCfg)
}

func UnmarshalRawConfig(buf []byte) (*ClashRawConfig, error) {
	// config with default value
	rawCfg := &ClashRawConfig{
		AllowLan:       false,
		BindAddress:    "*",
		Mode:           "rule",
		Authentication: []string{},
		LogLevel:       "info",
		Hosts:          map[string]string{},
		// Rule:           []string{},
		Proxy: []map[string]interface{}{},
		// ProxyGroup:     []map[string]interface{}{},
		Profile: Profile{
			StoreSelected: true,
		},
	}

	if err := yaml.Unmarshal(buf, &rawCfg); err != nil {
		return nil, err
	}

	return rawCfg, nil
}

func ParseRawConfig(rawCfg *ClashRawConfig) (*ClashConfig, error) {
	config := &ClashConfig{}

	config.Profile = &rawCfg.Profile

	general, err := parseGeneral(rawCfg)
	if err != nil {
		return nil, err
	}
	config.General = general

	proxies, err := parseProxies(rawCfg)
	if err != nil {
		return nil, err
	}
	config.Proxies = proxies

	return config, nil
}

func parseGeneral(cfg *ClashRawConfig) (*General, error) {
	return &General{
		Inbound: Inbound{
			Port:        cfg.Port,
			SocksPort:   cfg.SocksPort,
			RedirPort:   cfg.RedirPort,
			TProxyPort:  cfg.TProxyPort,
			MixedPort:   cfg.MixedPort,
			AllowLan:    cfg.AllowLan,
			BindAddress: cfg.BindAddress,
		},
		Controller: Controller{
			ExternalController: cfg.ExternalController,
			ExternalUI:         cfg.ExternalUI,
			Secret:             cfg.Secret,
		},
		Mode:      cfg.Mode,
		LogLevel:  cfg.LogLevel,
		IPv6:      cfg.IPv6,
		Interface: cfg.Interface,
	}, nil
}

func parseProxies(cfg *ClashRawConfig) ([]string, error) {

	proxyList := []string{}
	// parse proxy
	for idx, mapping := range cfg.Proxy {
		link, err := ParseProxy(mapping, cfg.NamePrefix)
		if err != nil {
			log.Printf("parseProxies %d: %s", idx, err.Error())
			continue
		}
		proxyList = append(proxyList, link)
	}
	return proxyList, nil
}

func hostWithDefaultPort(host string, defPort string) (string, error) {
	if !strings.Contains(host, ":") {
		host += ":"
	}

	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		return "", err
	}

	if port == "" {
		port = defPort
	}

	return net.JoinHostPort(hostname, port), nil
}

func parseNameServer(servers []string) ([]dns.NameServer, error) {
	nameservers := []dns.NameServer{}

	for idx, server := range servers {
		// parse without scheme .e.g 8.8.8.8:53
		if !strings.Contains(server, "://") {
			server = "udp://" + server
		}
		u, err := url.Parse(server)
		if err != nil {
			return nil, fmt.Errorf("DNS NameServer[%d] format error: %s", idx, err.Error())
		}

		var addr, dnsNetType string
		switch u.Scheme {
		case "udp":
			addr, err = hostWithDefaultPort(u.Host, "53")
			dnsNetType = "" // UDP
		case "tcp":
			addr, err = hostWithDefaultPort(u.Host, "53")
			dnsNetType = "tcp" // TCP
		case "tls":
			addr, err = hostWithDefaultPort(u.Host, "853")
			dnsNetType = "tcp-tls" // DNS over TLS
		case "https":
			clearURL := url.URL{Scheme: "https", Host: u.Host, Path: u.Path}
			addr = clearURL.String()
			dnsNetType = "https" // DNS over HTTPS
		default:
			return nil, fmt.Errorf("DNS NameServer[%d] unsupport scheme: %s", idx, u.Scheme)
		}

		if err != nil {
			return nil, fmt.Errorf("DNS NameServer[%d] format error: %s", idx, err.Error())
		}

		nameservers = append(
			nameservers,
			dns.NameServer{
				Net:  dnsNetType,
				Addr: addr,
			},
		)
	}
	return nameservers, nil
}

func parseFallbackIPCIDR(ips []string) ([]*net.IPNet, error) {
	ipNets := []*net.IPNet{}

	for idx, ip := range ips {
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, fmt.Errorf("DNS FallbackIP[%d] format error: %s", idx, err.Error())
		}
		ipNets = append(ipNets, ipnet)
	}

	return ipNets, nil
}
