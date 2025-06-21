package douyin

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 转换 JSON 到 Cookie 字符串
func JsonToCookieString(jsonString string) (string, error) {
	var cookies []Cookie
	// 解析 JSON
	err := json.Unmarshal([]byte(jsonString), &cookies)
	if err != nil {
		return "", err
	}
	// 构建 Cookie 字符串
	var cookieParts []string
	for _, cookie := range cookies {
		if cookie.Name != "" {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}
	return strings.Join(cookieParts, ";"), nil
}

// 获取 VerifyParams
func BuildVerifyParams(userAgent string) (VerifyParams, error) {
	tm := NewTokenManager(userAgent)
	vfm := VerifyFpManager{}

	msToken, _ := tm.GetMsToken()
	webid, _ := tm.GenWebID()
	verifyFp := vfm.GenVerifyFp()
	sVWebID := vfm.GenSVWebID()

	return VerifyParams{
		MsToken:  msToken,
		WebID:    webid,
		VerifyFp: verifyFp,
		SVWebID:  sVWebID,
	}, nil
}

// GetWebID 生成随机的 WebID
func GetWebID() string {
	rand.Seed(time.Now().UnixNano())

	// 生成类似 "10000000-1000-4000-8000-100000000000" 的随机 ID
	baseID := fmt.Sprintf("%d-%d-%d-%d-%d",
		int64(1e7),
		int64(1e3),
		int64(4e3),
		int64(8e3),
		int64(1e11),
	)

	// 生成最终的 WebID
	var webID strings.Builder
	for _, ch := range baseID {
		if ch == '-' {
			continue
		}
		num := int(ch - '0')
		if num == 0 || num == 1 || num == 8 {
			webID.WriteString(fmt.Sprintf("%d", num^((rand.Intn(16))>>(num/4))))
		} else {
			webID.WriteRune(ch)
		}
	}

	// 截取前 19 位
	return webID.String()[:19]
}

// 生成随机字符串
func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// GetCurrentMicroTimestamp 返回当前 Unix 时间戳（毫秒）
func GetCurrentMicroTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
