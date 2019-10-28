package elf

import (
	"encoding/base64"
	"strings"
)

// Base64SafeEncode 安全URL编码，去除后面多余的等号
func Base64SafeEncode(source []byte) string {
	dest := base64.URLEncoding.EncodeToString(source)

	return strings.TrimRight(dest, "=")
}

// Base64SafeDecode 安全解码，兼容标准和URL编码，以及后面等号是否多余
func Base64SafeDecode(source string) ([]byte, error) {
	src := source
	// Base64 Url Safe is the same as Base64 but does not contain '/' and '+'
	// (replaced by '_' and '-') and trailing '=' are removed.
	src = strings.Replace(src, "_", "/", -1)
	src = strings.Replace(src, "-", "+", -1)

	if i := len(src) % 4; i != 0 {
		src += strings.Repeat("=", 4-i)
	}

	return base64.StdEncoding.DecodeString(src)
}
