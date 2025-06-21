package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockServer 创建一个模拟的 HTTP 服务器
func mockServer(t *testing.T, response map[string]interface{}, statusCode int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})
	return httptest.NewServer(handler)
}

// TestHTTPClientRequest 测试 HTTP 请求
func TestHTTPClientRequest(t *testing.T) {
	mockResp := map[string]interface{}{"message": "ok"}
	server := mockServer(t, mockResp, http.StatusOK)
	defer server.Close()

	client := NewHttpClient(&HttpClientConfig{Timeout: 5})

	opts := &RequestOptions{
		Headers: map[string]string{"X-Test": "test"},
		Query:   map[string]string{"param": "123"},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Request("GET", server.URL, opts)

	assert.NoError(t, err, "请求应成功")
	assert.NotNil(t, resp.Body(), "响应不能为空")

	var data map[string]interface{}
	err = json.Unmarshal(resp.Body(), &data)
	assert.NoError(t, err, "解析 JSON 失败")
	assert.Equal(t, "ok", data["message"], "返回数据应匹配")
}
