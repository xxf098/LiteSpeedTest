package config

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/utils"
)

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
	Add        string          `json:"add"`
	Aid        json.RawMessage `json:"aid"`
	AlterId    json.RawMessage `json:"alterId,omitempty"`
	Host       string          `json:"host"`
	ID         string          `json:"id"`
	Net        string          `json:"net"`
	Path       string          `json:"path"`
	Port       json.RawMessage `json:"port"`
	Ps         string          `json:"ps"`
	TLS        string          `json:"tls"`
	Type       string          `json:"type"`
	V          json.RawMessage `json:"v,omitempty"`
	Security   string          `json:"security,omitempty"`
	Scy        string          `json:"scy,omitempty"`
	ResolveIP  bool            `json:"resolve_ip,omitempty"`
	ServerName string          `json:"-"`
	PortInt    int             `json:"-"`
	AidInt     int             `json:"-"`
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

func VmessConfigToVmessOption(config *VmessConfig) (*outbound.VmessOption, error) {
	port, err := rawMessageToInt(config.Port)
	if err != nil {
		return nil, err
	}
	aid, err := rawMessageToInt(config.Aid)
	if err != nil {
		return nil, err
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
		SkipCertVerify: false,
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
	}
	if config.Security != "" {
		vmessOption.Cipher = config.Security
	} else if config.Scy != "" {
		vmessOption.Cipher = config.Scy
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
	return VmessLinkToVmessOptionIP(link, false)
}

// TODO: safe base64
func VmessLinkToVmessOptionIP(link string, resolveip bool) (*outbound.VmessOption, error) {
	regex := regexp.MustCompile(`^vmess://([A-Za-z0-9+-=/_]+)`)
	res := regex.FindAllStringSubmatch(link, 1)
	b64 := ""
	if len(res) > 0 && len(res[0]) > 1 {
		b64 = res[0][1]
	}
	// data, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	if data, err = base64.RawStdEncoding.DecodeString(b64); err != nil {
	// 		return nil, err
	// 	}
	// }
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
	return VmessConfigToVmessOption(&config)
}

func VmessLinkToVmessConfigIP(link string, resolveip bool) (*VmessConfig, error) {
	regex := regexp.MustCompile(`^vmess://([A-Za-z0-9+-=/]+)`)
	res := regex.FindAllStringSubmatch(link, 1)
	b64 := ""
	if len(res) > 0 && len(res[0]) > 1 {
		b64 = res[0][1]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	// fmt.Println(string(data))
	if err != nil {
		return nil, err
	}
	config := VmessConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	port, err := rawMessageToInt(config.Port)
	if err != nil {
		return nil, err
	}
	config.PortInt = port
	aid, err := rawMessageToInt(config.Aid)
	if err != nil {
		aid, err = rawMessageToInt(config.AlterId)
		if err != nil {
			return nil, err
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
	return &config, nil
}

func ToVmessOption(path string) (*outbound.VmessOption, error) {
	data, err := ioutil.ReadFile(path)
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
