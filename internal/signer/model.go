package signer

import "encoding/json"

// XhsSignResult 代表小红书(Xhs)的签名结果
type XhsSignResult struct {
	XS         string `json:"x_s"`
	XT         string `json:"x_t"`
	XSCommon   string `json:"x_s_common"`
	XB3TraceID string `json:"x_b3_traceid"`
}

// XhsSignRequest 代表小红书(Xhs)的签名请求
type XhsSignRequest struct {
	URI     string      `json:"uri"`            // 请求的URI
	Data    interface{} `json:"data,omitempty"` // 请求body的数据
	Cookies string      `json:"cookies"`        // Cookies
}

// XhsSignResponse 代表小红书(Xhs)的签名响应
type XhsSignResponse struct {
	BizCode int            `json:"biz_code"`
	Msg     string         `json:"msg"`
	IsOK    bool           `json:"isok"`
	Data    *XhsSignResult `json:"data,omitempty"`
}

// DouyinSignResult 代表抖音(Douyin)的签名结果
type DouyinSignResult struct {
	ABogus string `json:"a_bogus"` // a_bogus签名
}

// DouyinSignRequest 代表抖音(Douyin)的签名请求
type DouyinSignRequest struct {
	URI         string `json:"uri"`          // 请求URI
	QueryParams string `json:"query_params"` // 请求的query_params (url encode 后的参数)
	UserAgent   string `json:"user_agent"`   // 请求的User-Agent
	Cookies     string `json:"cookies"`      // 请求的Cookies
}

// DouyinSignResponse 代表抖音(Douyin)的签名响应
type DouyinSignResponse struct {
	BizCode int               `json:"biz_code"`
	Msg     string            `json:"msg"`
	IsOK    bool              `json:"isok"`
	Data    *DouyinSignResult `json:"data,omitempty"`
}

// BilibiliSignResult 代表哔哩哔哩(Bilibili)的签名结果
type BilibiliSignResult struct {
	WTS  string `json:"wts"`   // 时间戳
	WRid string `json:"w_rid"` // 加密后的w_rid
}

// BilibiliSignRequest 代表哔哩哔哩(Bilibili)的签名请求
type BilibiliSignRequest struct {
	ReqData map[string]interface{} `json:"req_data"` // JSON格式的请求参数
	Cookies string                 `json:"cookies"`  // 登录成功后的Cookies
}

// BilibiliSignResponse 代表哔哩哔哩(Bilibili)的签名响应
type BilibiliSignResponse struct {
	BizCode int                 `json:"biz_code"`
	Msg     string              `json:"msg"`
	IsOK    bool                `json:"isok"`
	Data    *BilibiliSignResult `json:"data,omitempty"`
}

// ZhihuSignResult 代表知乎(Zhihu)的签名结果
type ZhihuSignResult struct {
	XZst81 string `json:"x_zst_81"`
	XZse96 string `json:"x_zse_96"`
}

// ZhihuSignRequest 代表知乎(Zhihu)的签名请求
type ZhihuSignRequest struct {
	URI     string `json:"uri"`     // 请求的URI
	Cookies string `json:"cookies"` // 请求的Cookies
}

// ZhihuSignResponse 代表知乎(Zhihu)的签名响应
type ZhihuSignResponse struct {
	BizCode int              `json:"biz_code"`
	Msg     string           `json:"msg"`
	IsOK    bool             `json:"isok"`
	Data    *ZhihuSignResult `json:"data,omitempty"`
}

// ToJSON 转换为 JSON
func ToJSON(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
