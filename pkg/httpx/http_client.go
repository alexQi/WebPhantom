package httpx

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"noctua/pkg/utils/str"
	"time"
)

// RequestOptions 允许调用方动态传递 HTTP 请求参数
type RequestOptions struct {
	Headers map[string]string
	Query   interface{}
	Payload interface{}
	Proxy   string
	Timeout time.Duration
}

type HttpClientConfig struct {
	Timeout time.Duration
}

// HttpClient 通用 HTTP 客户端
type HttpClient struct {
	client *resty.Client
	config *HttpClientConfig
}

// NewHttpClient 创建 HTTP 客户端
func NewHttpClient(config *HttpClientConfig) *HttpClient {
	return &HttpClient{
		config: config,
		client: resty.New().SetTimeout(config.Timeout),
	}
}

// SetQueryParams 构建参数字符串
func (h *HttpClient) SetQueryParams(req *resty.Request, data interface{}) error {
	params, err := str.StructToMap(data)
	if err != nil {
		return err
	}
	if len(params) > 0 {
		req.SetQueryParams(params)
	}
	return nil
}

// Request 发送 HTTP 请求，支持动态参数
func (h *HttpClient) Request(method, url string, opts *RequestOptions) (*resty.Response, error) {
	req := h.client.R()
	req.SetAuthScheme("")
	req.Header.Del("Accept")

	// 设置 Headers
	if opts.Headers != nil {
		req.SetHeaders(opts.Headers)
	}

	// 设置 Query 参数
	if opts.Query != nil {
		err := h.SetQueryParams(req, opts.Query)
		if err != nil {
			return nil, err
		}
	}

	// 设置 Body
	if opts.Payload != nil {
		req.SetBody(opts.Payload)
	}

	// 设置代理
	if opts.Proxy != "" {
		h.client.SetProxy(opts.Proxy)
	}

	// 设置超时
	if opts.Timeout > 0 {
		h.client.SetTimeout(opts.Timeout)
	}

	// 发送请求
	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %v", err)
	}

	// 检查状态码
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("HTTP 错误状态码: %d, 响应: %s", resp.StatusCode(), resp.String())
	}

	return resp, nil
}
