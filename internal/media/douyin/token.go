package douyin

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"strings"
	"time"
)

// TokenManager 负责 token 相关的管理
type TokenManager struct {
	userAgent string
}

// NewTokenManager 创建 TokenManager
func NewTokenManager(userAgent string) *TokenManager {
	return &TokenManager{
		userAgent: userAgent,
	}
}

// GetMsToken 获取 ms_token
func (t *TokenManager) GetMsToken() (string, error) {
	resp, err := resty.New().
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second).    // 初次重试等待 2s
		SetRetryMaxWaitTime(3 * time.Second). // 最大等待时间
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == 500 // 仅在 500 错误时重试
		}).
		R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json; charset=utf-8",
			"User-Agent":   t.userAgent,
		}).
		SetBody(map[string]interface{}{
			"magic":         538969122,
			"version":       1,
			"dataType":      8,
			"strData":       DOUYIN_MS_TOKEN_REQ_STR_DATA,
			"tspFromClient": time.Now().UnixMilli(),
			"ulr":           0,
		}).
		Post(DOUYIN_MS_TOKEN_REQ_URL)
	if err != nil {
		return "", fmt.Errorf("获取 ms_token 失败: %w", err)
	}
	cookies := resp.Cookies()
	// 解析cookies
	for _, cookie := range cookies {
		if cookie.Name == "msToken" {
			return cookie.Value, nil
			// 确保 msToken 长度符合要求
			//if len(cookie.Value) == 120 || len(cookie.Value) == 128 {
			//	return cookie.Value, nil
			//}
			//return "", fmt.Errorf("msToken 长度不符合要求: %d", len(cookie.Value))
		}
	}
	return "", fmt.Errorf("未找到有效的 msToken")
}

// GenWebID 获取 webid
func (t *TokenManager) GenWebID() (string, error) {
	var responseData map[string]interface{}
	resp, err := resty.New().
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second).    // 初次重试等待 2s
		SetRetryMaxWaitTime(3 * time.Second). // 最大等待时间
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == 500 // 仅在 500 错误时重试
		}).
		R().
		SetResult(&responseData).
		SetHeaders(map[string]string{
			"User-Agent":   t.userAgent,
			"Content-Type": "application/json; charset=UTF-8",
			"Referer":      "https://www.douyin.com/",
		}).
		SetBody(map[string]interface{}{
			"app_id":         6383,
			"referer":        "https://www.douyin.com/",
			"url":            "https://www.douyin.com/",
			"user_agent":     t.userAgent,
			"user_unique_id": "",
		}).
		Post(DOUYIN_WEBID_REQ_URL)

	if err != nil {
		return "", fmt.Errorf("获取 webid 失败: %w", err)
	}
	// **检查 JSON 解析是否成功**
	if resp.Result() == nil {
		return "", fmt.Errorf("JSON 解析失败，原始响应: %s", resp.String())
	}
	webID, ok := responseData["web_id"].(string)
	if !ok || webID == "" {
		return "", fmt.Errorf("webid 为空")
	}
	return webID, nil
}

// GenFakeMsToken 生成假 msToken
func (t *TokenManager) GenFakeMsToken() string {
	return GetRandomString(126) + "=="
}

// VerifyFpManager 生成 VerifyFp 和 s_v_web_id
type VerifyFpManager struct{}

// GenVerifyFp 生成 verifyFp
func (v VerifyFpManager) GenVerifyFp() string {
	const baseStr = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const t = len(baseStr)

	milliseconds := time.Now().UnixNano() / int64(time.Millisecond)
	base36 := ""
	for milliseconds > 0 {
		remainder := milliseconds % 36
		if remainder < 10 {
			base36 = fmt.Sprintf("%d%s", remainder, base36)
		} else {
			base36 = fmt.Sprintf("%c%s", 'a'+remainder-10, base36)
		}
		milliseconds /= 36
	}

	r := base36
	o := make([]string, 36)
	o[8], o[13], o[18], o[23] = "_", "_", "_", "_"
	o[14] = "4"

	for i := 0; i < 36; i++ {
		if o[i] == "" {
			n := rand.Intn(t)
			if i == 19 {
				n = (3 & n) | 8
			}
			o[i] = string(baseStr[n])
		}
	}

	return "verify_" + r + "_" + strings.Join(o, "")
}

// GenSVWebID 生成 s_v_web_id
func (v VerifyFpManager) GenSVWebID() string {
	return v.GenVerifyFp()
}
