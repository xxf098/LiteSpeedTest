package config

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseClash(t *testing.T) {
	dat, err := ioutil.ReadFile("./clash.yaml")
	if err != nil {
		t.Error(err)
	}
	c, err := ParseClash(dat)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(c.Proxies)
}

func TestShadowrocketLinkToVmessLink(t *testing.T) {
	link := "vmess://YXV0bzo0MzlkYzc0Yy02ZWQ5LTQ5MDQtODVjYi0yM2JlZTY1OGQ4Y2ZAanAyLm1heWl5dW4udmlwOjgw?tfo=1&remark=80%E4%B8%A8%E8%81%94%E9%80%9A%E6%89%8B%E5%8E%85%E4%B8%A8%E6%97%A5%E6%9C%AC1Gbps%E4%B8%A81&alterId=0&obfs=websocket&path=%2F&obfsParam=shoutingtoutiao3.10010.com"
	v2rayLink, err := ShadowrocketLinkToVmessLink(link)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v2rayLink)
}
