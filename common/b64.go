package common

import (
	"encoding/base64"
	"strings"
	"unsafe"
)

// unsafe
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
	return b2s(data), nil
}

// nolint
func b2s(b []byte) string { return *(*string)(unsafe.Pointer(&b)) } // tricks

func DecodeB64Bytes(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		if data, err = base64.RawStdEncoding.DecodeString(s); err != nil {
			return nil, err
		}
	}
	return data, nil
}
