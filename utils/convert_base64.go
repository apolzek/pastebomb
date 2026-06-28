package utils

import "encoding/base64"

func EncodeContent(content string) *string {
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(content))
	return &contentBase64
}
