package web

import (
	"encoding/json"
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
	Group    string `json:"group"`
	Ping     int64  `json:"ping"`
	Lost     string `json:"lost"`
	Speed    string `json:"speed"`
	MaxSpeed string `json:"maxspeed"`
	Link     string `json:"link"`
}

func getRemarks(link string) (string, error) {
	cfgVmess, err := config.VmessLinkToVmessConfigIP(link, false)
	if err == nil {
		return cfgVmess.Ps, nil
	}
	cfgSSR, err := config.SSRLinkToSSROption(link)
	if err == nil {
		return cfgSSR.Remarks, nil
	}
	cfgTrojan, err := config.TrojanLinkToTrojanOption(link)
	if err == nil {
		return cfgTrojan.Remarks, nil
	}
	cfgSS, err := config.SSLinkToSSOption(link)
	if err == nil {
		return cfgSS.Remarks, nil
	}
	return "", nil
}

func gotserverMsg(id int, link string) []byte {
	msg := Message{ID: id, Info: "gotserver"}
	remarks, err := getRemarks(link)
	if err == nil {
		msg.Group = "Group 1"
		msg.Remarks = remarks
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
		var ping int64
		if len(option) > 0 {
			if v, ok := option[0].(int64); ok {
				ping = v
			}
		}
		msg.Ping = ping
	case "gotspeed":
		var speed int64
		var maxspeed int64
		if len(option) > 1 {
			if v, ok := option[0].(int64); ok {
				speed = v
			}
			if v, ok := option[1].(int64); ok {
				maxspeed = v
			}
		}
		msg.Speed = strings.TrimRight(download.ByteCountIEC(speed), "/s")
		msg.MaxSpeed = strings.TrimRight(download.ByteCountIEC(maxspeed), "/s")
		if speed < 1 {
			msg.Speed = "N/A"
		}
		if maxspeed < 1 {
			msg.MaxSpeed = "N/A"
		}
	}
	b, _ := json.Marshal(msg)
	return b
}
