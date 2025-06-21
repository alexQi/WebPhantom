package httpx

import (
	"net/url"
	"strings"
)

// EncodeURLParams 将 map[string]string 转换为 URL 查询字符串
func EncodeURLParams(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return strings.TrimPrefix(values.Encode(), "?")
}
