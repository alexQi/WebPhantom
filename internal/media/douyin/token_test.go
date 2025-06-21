package douyin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetMsToken 测试获取 msToken
func TestGetMsToken(t *testing.T) {
	tokenManager := NewTokenManager(DOUYIN_FIXED_USER_AGENT)

	msToken, err := tokenManager.GetMsToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, msToken)
	assert.Contains(t, []int{120, 128}, len(msToken), "msToken 长度应为 120 或 128")
}

// TestGenFakeMsToken 测试生成假的 msToken
func TestGenFakeMsToken(t *testing.T) {
	tokenManager := NewTokenManager(DOUYIN_FIXED_USER_AGENT)

	msToken := tokenManager.GenFakeMsToken()
	assert.NotEmpty(t, msToken)
	assert.Len(t, msToken, 128, "伪造 msToken 长度应为 128")
}

// TestGenWebID 测试获取 webID
func TestGenWebID(t *testing.T) {
	tokenManager := NewTokenManager(DOUYIN_FIXED_USER_AGENT)

	webID, err := tokenManager.GenWebID()

	assert.NoError(t, err, "应返回没有错误")
	assert.NotEmpty(t, webID, "webID 不为空")
}

// TestGenVerifyFp 测试生成 verifyFp
func TestGenVerifyFp(t *testing.T) {
	manager := VerifyFpManager{}
	verifyFp := manager.GenVerifyFp()
	assert.NotEmpty(t, verifyFp)
	assert.Contains(t, verifyFp, "verify_", "verifyFp 应该以 'verify_' 开头")
	assert.Len(t, verifyFp, 52, "verifyFp 长度应为 44")
}

// TestGenSVWebID 测试生成 s_v_web_id
func TestGenSVWebID(t *testing.T) {
	manager := VerifyFpManager{}
	sVWebID := manager.GenSVWebID()
	assert.NotEmpty(t, sVWebID)
	assert.Contains(t, sVWebID, "verify_", "s_v_web_id 应该以 'verify_' 开头")
	assert.Len(t, sVWebID, 52, "s_v_web_id 长度应为 44")
}
