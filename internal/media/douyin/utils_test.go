package douyin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWebID(t *testing.T) {
	webID := GetWebID()
	assert.Len(t, webID, 19, "WebID 长度应为 19")
}

func TestGetRandomString(t *testing.T) {
	randomStr := GetRandomString(16)
	assert.Len(t, randomStr, 16, "随机字符串长度应为 16")

	randomStr2 := GetRandomString(32)
	assert.Len(t, randomStr2, 32, "随机字符串长度应为 32")
}

func TestGetCurrentMicroTimestamp(t *testing.T) {
	timestamp := GetCurrentMicroTimestamp()
	assert.Greater(t, timestamp, int64(0), "时间戳应为正数")

	// 确保时间戳在合理范围内（例如，不是 0，也不是很小的值）
	now := time.Now().UnixNano() / int64(time.Millisecond)
	assert.Greater(t, timestamp, now-int64(1000), "时间戳应接近当前时间")
	assert.Less(t, timestamp, now+int64(1000), "时间戳应接近当前时间")
}

func TestBuildVerifyParams(t *testing.T) {
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	params, err := BuildVerifyParams(userAgent)

	assert.NoError(t, err, "获取 VerifyParams 不应报错")
	assert.NotEmpty(t, params.MsToken, "msToken 不应为空")
	assert.NotEmpty(t, params.WebID, "webID 不应为空")
	assert.NotEmpty(t, params.VerifyFp, "verifyFp 不应为空")
	assert.NotEmpty(t, params.SVWebID, "sVWebID 不应为空")
}
