package signer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const SIGN_SERVER = "http://127.0.0.1:8989"

// TestPongSignServer 测试签名服务器 Pong
func TestPongSignServer(t *testing.T) {
	client := NewSignServerClient(SIGN_SERVER)
	err := client.PongSignServer()

	assert.NoError(t, err, "PongSignServer 应该成功返回")
}

// TestDouyinSign 测试抖音签名
func TestDouyinSign(t *testing.T) {
	client := NewSignServerClient(SIGN_SERVER)
	req := &DouyinSignRequest{URI: "/test"}

	resp, err := client.DouyinSign(req)

	assert.NoError(t, err, "请求应该成功")
	assert.NotNil(t, resp, "响应不能为空")
	assert.NotEmpty(t, resp.Data.ABogus, "a_bogus 不能为空")
}

// TestBilibiliSign 测试 Bilibili 签名
func TestBilibiliSign(t *testing.T) {
	client := NewSignServerClient(SIGN_SERVER)
	req := &BilibiliSignRequest{
		ReqData: make(map[string]interface{}),
	}

	resp, err := client.BilibiliSign(req)

	assert.NoError(t, err, "请求应该成功")
	assert.NotNil(t, resp, "响应不能为空")
	assert.NotEmpty(t, resp.Data.WTS, "wts 不能为空")
	assert.NotEmpty(t, resp.Data.WRid, "w_rid 不能为空")
}

// TestZhihuSign 测试 知乎 签名
func TestZhihuSign(t *testing.T) {
	client := NewSignServerClient(SIGN_SERVER)
	req := &ZhihuSignRequest{}

	resp, err := client.ZhihuSign(req)

	assert.NoError(t, err, "请求应该成功")
	assert.NotNil(t, resp, "响应不能为空")
	assert.NotEmpty(t, resp.Data.XZst81, "x_zst_81 不能为空")
	assert.NotEmpty(t, resp.Data.XZse96, "x_zse_96 不能为空")
}

// TestXiaohongshuSign 测试 小红书 签名
func TestXiaohongshuSign(t *testing.T) {
	client := NewSignServerClient(SIGN_SERVER)
	req := &XhsSignRequest{}

	resp, err := client.XiaohongshuSign(req)

	assert.NoError(t, err, "请求应该成功")
	assert.NotNil(t, resp, "响应不能为空")
	assert.NotEmpty(t, resp.Data.XS, "x_s 不能为空")
	assert.NotEmpty(t, resp.Data.XT, "x_t 不能为空")
	assert.NotEmpty(t, resp.Data.XSCommon, "x_s_common 不能为空")
	assert.NotEmpty(t, resp.Data.XB3TraceID, "x_b3_traceid 不能为空")
}
