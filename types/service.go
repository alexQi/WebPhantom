package types

// ProxyRequest 代理请求
type ProxyRequest struct {
	Num      int    `json:"num"`
	Type     string `json:"type"`
	Region   string `json:"region"`
	ProxyKey string `json:"proxyKey"`
}

// ProxyResponse API 返回结构
type ProxyResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Mesc int         `json:"mesc"`
	Data []ProxyInfo `json:"data"`
}

type ReleaseData struct {
	Code      string   `json:"code"`
	Version   string   `json:"version"`
	Download  string   `json:"download"`
	Changelog []string `json:"changelog"`
}

type PongResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Mesc int         `json:"mesc"`
	Data ReleaseData `json:"data"`
}
