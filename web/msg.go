package web

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/download"
)

const (
	TROJAN_DEFAULT_GROUP    = "TrojanProvider"
	V2RAY_DEFAULT_GROUP     = "V2RayProvider"
	SS_DEFAULT_GROUP        = "SSProvider"
	SSR_DEFAULT_GROUP       = "SSRProvider"
	SPEEDTEST_ERROR_NONODES = "{\"info\":\"error\",\"reason\":\"nonodes\"}\n"
)

type Message struct {
	ID       int       `json:"id"`
	Info     string    `json:"info"`
	Remarks  string    `json:"remarks,omitempty"`
	Server   string    `json:"server,omitempty"`
	Group    string    `json:"group,omitempty"`
	Ping     int64     `json:"ping,omitempty"`
	Lost     string    `json:"lost,omitempty"`
	Speed    string    `json:"speed,omitempty"`
	MaxSpeed string    `json:"maxspeed,omitempty"`
	Traffic  int64     `json:"traffic,omitempty"`
	Link     string    `json:"link,omitempty"`
	Protocol string    `json:"protocol,omitempty"`
	PicData  string    `json:"data,omitempty"`
	Servers  []Message `json:"servers,omitempty"`
}

func GetRemarks(link string) (string, string, error) {
	cfg, err := config.Link2Config(link)
	if err != nil {
		return "", "", err
	}
	return cfg.Protocol, cfg.Remarks, err
}

func gotserverMsg(id int, link string, groupName string) []byte {
	msg := Message{ID: id, Info: "gotserver"}
	cfg, err := config.Link2Config(link)
	if err == nil {
		msg.Group = groupName
		msg.Remarks = cfg.Remarks
		msg.Server = net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port))
		msg.Protocol = cfg.Protocol
		if cfg.Protocol == "vmess" && cfg.Net != "" {
			msg.Protocol = fmt.Sprintf("%s/%s", cfg.Protocol, cfg.Net)
		}
		msg.Link = link
	}
	b, _ := json.Marshal(msg)
	return b
}

func gotserversMsg(startID int, links []string, groupName string) []byte {
	servers := Message{ID: startID, Info: "gotservers"}
	for i, link := range links {
		id := startID + i
		msg := Message{ID: id, Info: "gotserver"}
		cfg, err := config.Link2Config(link)
		if err == nil {
			msg.Group = groupName
			msg.Remarks = cfg.Remarks
			msg.Server = net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port))
			msg.Protocol = cfg.Protocol
			if (cfg.Protocol == "vmess" || cfg.Protocol == "trojan") && cfg.Net != "" {
				msg.Protocol = fmt.Sprintf("%s/%s", cfg.Protocol, cfg.Net)
			}
			msg.Link = link
		}
		servers.Servers = append(servers.Servers, msg)
	}
	b, _ := json.Marshal(servers)
	return b
}

func getMsgByte(id int, typ string, option ...interface{}) []byte {
	msg := Message{ID: id, Info: typ}
	switch typ {
	case "gotping":
		msg.Lost = "0.00%"
		if len(option) > 0 {
			if v, ok := option[0].(int64); ok {
				msg.Ping = v
			}
		}

	case "gotspeed":
		var speed int64
		var maxspeed int64
		var traffic int64
		if len(option) > 1 {
			if v, ok := option[0].(int64); ok {
				speed = v
			}
			if v, ok := option[1].(int64); ok {
				maxspeed = v
			}
			if len(option) > 2 {
				if v, ok := option[2].(int64); ok {
					traffic = v
				}
			}
		}
		msg.Speed = strings.TrimRight(download.ByteCountIEC(speed), "/s")
		msg.MaxSpeed = strings.TrimRight(download.ByteCountIEC(maxspeed), "/s")
		msg.Traffic = traffic
		if speed < 1 {
			msg.Speed = "N/A"
		}
		if maxspeed < 1 {
			msg.MaxSpeed = "N/A"
		}
	case "picdata":
		if len(option) > 0 {
			if v, ok := option[0].(string); ok {
				msg.PicData = v
			}
		}
	}
	b, _ := json.Marshal(msg)
	return b
}
