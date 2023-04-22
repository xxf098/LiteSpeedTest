package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/utils"
)

var RegShadowrocketVmess = regexp.MustCompile(`(?i)vmess://(\S+?)@(\S+?):([0-9]{2,5})/?([?#][^\s]+)`)

type User struct {
	Email    string `json:"Email"`
	ID       string `json:"ID"`
	AlterId  int    `json:"alterId"`
	Security string `json:"security"`
}

type VNext struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	Users   []User `json:"users"`
}

type Settings struct {
	Vnexts []VNext `json:"vnext"`
}

type WSSettings struct {
	Path string `json:"path"`
}

type StreamSettings struct {
	Network    string     `json:"network"`
	Security   string     `json:"security"`
	WSSettings WSSettings `json:"wsSettings,omitempty"`
}

type Outbound struct {
	Protocol       string          `json:"protocol"`
	Description    string          `json:"description"`
	Settings       Settings        `json:"settings"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
}

type RawConfig struct {
	Outbounds []Outbound `json:"outbounds"`
}

type VmessConfig struct {
	Add            string          `json:"add"`
	Aid            json.RawMessage `json:"aid"`
	AlterId        json.RawMessage `json:"alterId,omitempty"`
	Host           string          `json:"host"`
	ID             string          `json:"id"`
	Net            string          `json:"net"`
	Path           string          `json:"path"`
	Port           json.RawMessage `json:"port"`
	Ps             string          `json:"ps"`
	TLSRaw         json.RawMessage `json:"tls"`
	Type           string          `json:"type"`
	V              json.RawMessage `json:"v,omitempty"`
	Security       string          `json:"security,omitempty"`
	Scy            string          `json:"scy,omitempty"`
	Encryption     string          `json:"encryption,omitempty"`
	ResolveIP      bool            `json:"resolve_ip,omitempty"`
	SkipCertVerify bool            `json:"skip-cert-verify,omitempty"`
	ServerName     string          `json:"sni,omitempty"`
	PortInt        int             `json:"-"`
	AidInt         int             `json:"-"`
	TLS            string          `json:"-"`
}

type VmessConfigMarshal struct {
	Add            string `json:"add"`
	Aid            int    `json:"aid"`
	Host           string `json:"host"`
	ID             string `json:"id"`
	Net            string `json:"net"`
	Path           string `json:"path"`
	Port           uint16 `json:"port"`
	Ps             string `json:"ps"`
	TLS            string `json:"tls"`
	Type           string `json:"type"`
	V              string `json:"v,omitempty"`
	Security       string `json:"security,omitempty"`
	Scy            string `json:"scy,omitempty"`
	ResolveIP      bool   `json:"resolve_ip,omitempty"`
	SkipCertVerify bool   `json:"skip-cert-verify"`
	ServerName     string `json:"sni"`
}

func RawConfigToVmessOption(config *RawConfig) (*outbound.VmessOption, error) {
	var ob Outbound
	for _, outbound := range config.Outbounds {
		if outbound.Protocol == "vmess" {
			ob = outbound
			break
		}
	}
	vnext := ob.Settings.Vnexts[0]
	vmessOption := outbound.VmessOption{
		HTTPOpts: outbound.HTTPOptions{
			Method: "GET",
			Path:   []string{"/"},
		},
		Name:           "vmess",
		Server:         vnext.Address,
		Port:           vnext.Port,
		UUID:           vnext.Users[0].ID,
		AlterID:        vnext.Users[0].AlterId,
		Cipher:         vnext.Users[0].Security,
		TLS:            false,
		UDP:            false,
		Network:        "tcp",
		SkipCertVerify: false,
	}
	if ob.StreamSettings != nil {
		if ob.StreamSettings.Security == "tls" {
			vmessOption.TLS = true
		}
		if ob.StreamSettings.Network == "ws" {
			vmessOption.Network = "ws"
			vmessOption.WSPath = ob.StreamSettings.WSSettings.Path
			if ob.StreamSettings.WSSettings.Path != "" {
				vmessOption.WSHeaders = map[string]string{
					"Host": vnext.Address,
				}
			}
		}
	}
	return &vmessOption, nil
}
func rawMessageToInt(raw json.RawMessage) (int, error) {
	var i int
	err := json.Unmarshal(raw, &i)
	if err != nil {
		var s string
		err := json.Unmarshal(raw, &s)
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(s)
	}
	return i, nil
}

func rawMessageToTLS(raw json.RawMessage) (string, error) {
	var s string
	err := json.Unmarshal(raw, &s)
	if err != nil {
		var b bool
		err := json.Unmarshal(raw, &b)
		if err != nil {
			return "", err
		}
		if b {
			s = "tls"
		}
	}
	return s, nil
}

func checkCipher(cipher string) bool {
	return cipher == "auto" || cipher == "none" || cipher == "aes-128-gcm" || cipher == "chacha20-poly1305"
}

func VmessConfigToVmessOption(config *VmessConfig) (*outbound.VmessOption, error) {
	port, err := rawMessageToInt(config.Port)
	if err != nil {
		port = 443
	}
	aid, err := rawMessageToInt(config.Aid)
	if err != nil {
		aid = 0
	}

	vmessOption := outbound.VmessOption{
		// HTTPOpts: outbound.HTTPOptions{
		// 	Method: "GET",
		// 	Path:   []string{"/"},
		// },
		Name:           "vmess",
		Server:         config.Add,
		Port:           uint16(port),
		UUID:           config.ID,
		AlterID:        aid,
		Cipher:         "none",
		TLS:            false,
		UDP:            false,
		Network:        "tcp",
		SkipCertVerify: config.SkipCertVerify,
		Type:           config.Type,
	}
	// http network
	if config.Type == "http" {
		vmessOption.HTTPOpts = outbound.HTTPOptions{
			Method: "GET",
			Path:   []string{config.Path},
		}
		if config.Host != "" {
			vmessOption.HTTPOpts.Headers = map[string][]string{
				"Host":       {config.Host},
				"Connection": {"keep-alive"},
			}
		}
		vmessOption.Network = "http"
	}
	if config.ResolveIP {
		if ipAddr, err := resolveIP(vmessOption.Server); err == nil && ipAddr != "" {
			vmessOption.ServerName = vmessOption.Server
			vmessOption.Server = ipAddr
		}
	}
	if config.TLS == "tls" {
		vmessOption.TLS = true
		if len(config.ServerName) > 0 && config.ServerName != config.Add {
			config.SkipCertVerify = true
		}
	}
	// check cipher
	if checkCipher(config.Security) {
		vmessOption.Cipher = config.Security
	} else if checkCipher(config.Scy) {
		vmessOption.Cipher = config.Scy
	} else if checkCipher(config.Encryption) {
		vmessOption.Cipher = config.Encryption
	}
	if config.Net == "ws" {
		vmessOption.Network = "ws"
		vmessOption.WSPath = config.Path
		vmessOption.WSHeaders = map[string]string{
			"Host": config.Host,
		}
	}
	if config.Net == "h2" {
		vmessOption.Network = "h2"
		if vmessOption.TLS {
			vmessOption.SkipCertVerify = false
		}
		vmessOption.HTTP2Opts = outbound.HTTP2Options{
			Host: []string{config.Host},
			Path: config.Path,
		}
	}
	return &vmessOption, nil
}

func VmessLinkToVmessOption(link string) (*outbound.VmessOption, error) {
	opt, err := VmessLinkToVmessOptionIP(link, false)
	if err != nil {
		return ShadowrocketVmessLinkToVmessOptionIP(link, false)
	}
	return opt, nil
}

// TODO: safe base64
func VmessLinkToVmessOptionIP(link string, resolveip bool) (*outbound.VmessOption, error) {
	config, err := VmessLinkToVmessConfig(link, resolveip)
	if err != nil {
		return nil, err
	}
	return VmessConfigToVmessOption(config)
}

func VmessLinkToVmessConfig(link string, resolveip bool) (*VmessConfig, error) {
	// FIXME:
	regex := regexp.MustCompile(`^vmess://([A-Za-z0-9+-=/_]+)`)
	res := regex.FindAllStringSubmatch(link, 1)
	b64 := ""
	if len(res) > 0 && len(res[0]) > 1 {
		b64 = res[0][1]
	}
	data, err := utils.DecodeB64Bytes(b64)
	if err != nil {
		return nil, err
	}
	config := VmessConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	config.ResolveIP = resolveip
	// parse raw message
	if tls, err := rawMessageToTLS(config.TLSRaw); err == nil {
		config.TLS = tls
	}
	return &config, nil
}

// parse shadowrocket link
func ShadowrocketVmessLinkToVmessOptionIP(link string, resolveip bool) (*outbound.VmessOption, error) {
	config, err := ShadowrocketVmessLinkToVmessConfig(link, resolveip)
	if err != nil {
		return nil, err
	}
	return VmessConfigToVmessOption(config)
}

func ShadowrocketLinkToVmessLink(link string) (string, error) {
	config, err := ShadowrocketVmessLinkToVmessConfig(link, false)
	if err != nil {
		return "", err
	}
	src, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("vmess://%s", base64.StdEncoding.EncodeToString(src)), nil
}

func ShadowrocketVmessLinkToVmessConfig(link string, resolveip bool) (*VmessConfig, error) {
	if !RegShadowrocketVmess.MatchString(link) {
		return nil, fmt.Errorf("not a vmess link: %s", link)
	}
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	config := VmessConfig{}
	config.V = []byte("2")

	b64 := url.Host
	b, err := utils.DecodeB64Bytes(b64)
	if err != nil {
		return shadowrocketVmessURLToVmessConfig(link, resolveip)
	}

	mhp := strings.SplitN(string(b), ":", 3)
	if len(mhp) != 3 {
		return nil, fmt.Errorf("vmess unrecognized: method:host:port -- %v", mhp)
	}
	config.Security = mhp[0]
	// mhp[0] is the encryption method
	config.Port = []byte(mhp[2])
	idadd := strings.SplitN(mhp[1], "@", 2)
	if len(idadd) != 2 {
		return nil, fmt.Errorf("vmess unrecognized: id@addr -- %v", idadd)
	}
	config.ID = idadd[0]
	config.Add = idadd[1]
	config.Aid = []byte("0")

	vals := url.Query()
	if v := vals.Get("remarks"); v != "" {
		config.Ps = v
	}
	if v := vals.Get("path"); v != "" {
		config.Path = v
	}
	if v := vals.Get("tls"); v == "1" {
		config.TLS = "tls"
	}
	if v := vals.Get("alterId"); v != "" {
		config.Aid = []byte(v)
		config.AlterId = []byte(v)
	}
	if v := vals.Get("obfs"); v != "" {
		switch v {
		case "websocket":
			config.Net = "ws"
		case "none":
			config.Net = "tcp"
			config.Type = "none"
		}
	}
	if v := vals.Get("obfsParam"); v != "" {
		config.Host = v
	}
	config.ResolveIP = resolveip
	return &config, nil
}

func shadowrocketVmessURLToVmessConfig(link string, resolveip bool) (*VmessConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "vmess" {
		return nil, fmt.Errorf("not a vmess link: %s", link)
	}
	pass := u.User.Username()
	hostport := u.Host
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, err
	}
	_, err = strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	config := VmessConfig{
		V:      []byte("2"),
		ID:     pass,
		Add:    host,
		Port:   []byte(u.Port()),
		Ps:     u.Fragment,
		Net:    "tcp",
		Aid:    []byte("0"),
		Type:   "auto",
		TLSRaw: []byte("false"),
	}

	vals, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	if v := vals.Get("type"); v != "" {
		config.Net = v
	}

	if v := vals.Get("encryption"); v != "" {
		config.Encryption = v
	}

	if v := vals.Get("host"); v != "" {
		config.Host = v
	}

	if v := vals.Get("path"); v != "" {
		config.Path = v
	}

	if v := vals.Get("security"); v != "" {
		config.TLS = v
		config.TLSRaw = json.RawMessage("true")
	}

	if v := vals.Get("sni"); v != "" {
		config.ServerName = v
	}

	if v := vals.Get("aid"); v != "" {
		config.Aid = []byte(v)
	}

	config.ResolveIP = resolveip
	return &config, nil
}

func VmessLinkToVmessConfigIP(link string, resolveip bool) (*VmessConfig, error) {
	var config *VmessConfig
	var err error
	if strings.Contains(link, "&") {
		config, err = ShadowrocketVmessLinkToVmessConfig(link, resolveip)
	} else {
		config, err = VmessLinkToVmessConfig(link, resolveip)
	}
	if err != nil {
		return nil, err
	}
	port, err := rawMessageToInt(config.Port)
	if err != nil {
		port = 443
	}
	config.PortInt = port
	aid, err := rawMessageToInt(config.Aid)
	if err != nil {
		aid, err = rawMessageToInt(config.AlterId)
		if err != nil {
			aid = 0
		}
	}
	config.AidInt = aid
	if resolveip {
		if ipAddr, err := resolveIP(config.Add); err == nil && ipAddr != "" {
			config.ServerName = config.Add
			if config.Host == "" {
				config.Host = config.ServerName
			}
			config.Add = ipAddr

		}
	}
	return config, nil
}

func ToVmessOption(path string) (*outbound.VmessOption, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := RawConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	if config.Outbounds != nil {
		return RawConfigToVmessOption(&config)
	}
	config1 := VmessConfig{}
	err = json.Unmarshal(data, &config1)
	if err != nil {
		return nil, err
	}
	return VmessConfigToVmessOption(&config1)
}

func init() {
	outbound.RegisterDialerCreator("vmess", func(link string) (outbound.Dialer, error) {
		vmessOption, err := VmessLinkToVmessOption(link)
		if err != nil {
			return nil, err
		}
		return outbound.NewVmess(vmessOption)
	})
}
