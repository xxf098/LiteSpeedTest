package common

import (
	"encoding/base64"
	"strings"
)

func DecodeB64(s string) (string, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		if data, err = base64.RawStdEncoding.DecodeString(s); err != nil {
			return "", err
		}
	}
	return string(data), nil
}
