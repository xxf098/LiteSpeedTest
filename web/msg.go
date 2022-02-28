package web

import (
	"encoding/json"
	"fmt"
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
	ID       int    `json:"id"`
	Info     string `json:"info"`
	Remarks  string `json:"remarks"`
	Server   string `json:"server"`
	Group    string `json:"group"`
	Ping     int64  `json:"ping"`
	Lost     string `json:"lost"`
	Speed    string `json:"speed"`
	MaxSpeed string `json:"maxspeed"`
	Traffic  int64  `json:"traffic"`
	Link     string `json:"link"`
	Protocol string `json:"protocol"`
	PicData  string `json:"data"`
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
		msg.Server = fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
		msg.Protocol = cfg.Protocol
		if cfg.Protocol == "vmess" && cfg.Net != "" {
			msg.Protocol = fmt.Sprintf("%s/%s", cfg.Protocol, cfg.Net)
		}
		msg.Link = link
	}
	b, _ := json.Marshal(msg)
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
