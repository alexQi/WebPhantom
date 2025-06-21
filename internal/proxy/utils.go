package proxy

import (
	"github.com/go-resty/resty/v2"
	"noctua/pkg/logger"
	"noctua/types"
	"time"
)

const TEST_PROXY_IP_URL = "https://cip.cc"

func isValidProxy(proxyInfo *types.ProxyInfo) (bool, error) {
	if proxyInfo.GetExpireTime().After(time.Now()) {
		return true, nil
	}
	resp, err := resty.New().
		SetTimeout(5 * time.Second).
		SetProxy(proxyInfo.BuildProtocol()).
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(3 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == 500
		}).
		R().
		Get(TEST_PROXY_IP_URL)
	if err != nil {
		logger.Log.Errorf("Checking proxy IP failed: %s, err: %v", proxyInfo.ProxyVal, err)
		return false, err
	}
	logger.Log.Infof("Checking proxy IP success: %s", proxyInfo.ProxyVal)
	return resp.StatusCode() == 200, nil
}
