package storage

import (
	"encoding/base64"
	"encoding/hex"
)

func base64Encode(src string) string {
	codec := base64.StdEncoding
	return codec.EncodeToString([]byte(src))
}

func hexEncode(src string) string {
	return hex.EncodeToString([]byte(src))
}
