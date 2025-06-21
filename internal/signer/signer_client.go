package signer

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

// SignServerClient 负责与签名服务器通信
type SignServerClient struct {
	Endpoint   string
	HttpClient *resty.Client
}

// NewSignServerClient 创建新的 SignServerClient
func NewSignServerClient(endpoint string) *SignServerClient {
	return &SignServerClient{
		Endpoint:   endpoint,
		HttpClient: resty.New().SetBaseURL(endpoint),
	}
}

// XiaohongshuSign 发送小红书签名请求
func (s *SignServerClient) XiaohongshuSign(reqData *XhsSignRequest) (*XhsSignResponse, error) {
	result := &XhsSignResponse{}
	_, err := s.HttpClient.R().SetResult(result).SetBody(reqData).Post("/signsrv/v1/xhs/sign")
	if err != nil {
		return nil, err
	}
	return result, err
}

// DouyinSign 发送抖音签名请求
func (s *SignServerClient) DouyinSign(reqData *DouyinSignRequest) (*DouyinSignResponse, error) {
	result := &DouyinSignResponse{}
	_, err := s.HttpClient.R().SetResult(result).SetBody(reqData).Post("/signsrv/v1/douyin/sign")
	if err != nil {
		return nil, err
	}
	return result, err
}

// BilibiliSign 发送哔哩哔哩签名请求
func (s *SignServerClient) BilibiliSign(reqData *BilibiliSignRequest) (*BilibiliSignResponse, error) {
	result := &BilibiliSignResponse{}
	_, err := s.HttpClient.R().SetResult(result).SetBody(reqData).Post("/signsrv/v1/bilibili/sign")
	if err != nil {
		return nil, err
	}
	return result, err
}

// ZhihuSign 发送知乎签名请求
func (s *SignServerClient) ZhihuSign(reqData *ZhihuSignRequest) (*ZhihuSignResponse, error) {
	result := &ZhihuSignResponse{}
	_, err := s.HttpClient.R().SetResult(result).SetBody(reqData).Post("/signsrv/v1/zhihu/sign")
	if err != nil {
		return nil, err
	}
	return result, err
}

// PongSignServer 检测签名服务器是否可用
func (s *SignServerClient) PongSignServer() error {
	resp, err := s.HttpClient.
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			fmt.Println("✅ 请求服务器成功，状态码:", r.StatusCode())
			return nil
		}).
		R().
		Get("/signsrv/pong")
	if err != nil {
		return fmt.Errorf("签名服务器不可用: %v", err)
	}

	fmt.Println("✅ 签名服务器正常运行:", string(resp.Body()))

	return nil
}
