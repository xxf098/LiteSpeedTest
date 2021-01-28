package utils

import (
	"regexp"

	"github.com/xxf098/lite-proxy/common"
)

func CheckLink(link string) ([]string, error) {
	r := regexp.MustCompile("(?i)^(vmess|trojan|ss|ssr)://.+")
	matches := r.FindStringSubmatch(link)
	if len(matches) < 2 {
		return nil, common.NewError("Not Suported Link")
	}
	return matches, nil
}
