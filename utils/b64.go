package utils

import (
	"encoding/base64"
	"strings"
)

func DecodeB64(s string) (string, error) {
	// s = strings.ReplaceAll(s, "-", "+")
	// s = strings.ReplaceAll(s, "_", "/")
	// data, err := base64.StdEncoding.DecodeString(s)
	// if err != nil {
	// 	if data, err = base64.RawStdEncoding.DecodeString(s); err != nil {
	// 		return "", err
	// 	}
	// }
	// return string(data), nil
	data, err := DecodeB64Bytes(s)
	if err != nil {
		return "", err
	}
	return B2s(data), nil
}

func DecodeB64Bytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if pad := len(s) % 4; pad != 0 {
		s += strings.Repeat("=", 4-pad)
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// URLEncoding
		if data, err = base64.RawStdEncoding.DecodeString(s); err != nil {
			return nil, err
		}
	}
	return data, nil
}
