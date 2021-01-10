package common

import "encoding/base64"

func DecodeB64(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		if data, err = base64.RawStdEncoding.DecodeString(s); err != nil {
			return "", err
		}
	}
	return string(data), nil
}
