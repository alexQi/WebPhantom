package types

import (
	"fmt"
	"time"
)

// IpInfo 代理 IP 结构体
type ProxyInfo struct {
	Useable    bool   `json:"useable"`    // 是否可用
	ProxyType  string `json:"proxyType"`  // 代理类型
	ProxyKey   string `json:"proxyKey"`   // 代理存储的key
	Protocol   string `json:"protocol"`   // 协议 (http/https)
	ProxyVal   string `json:"proxyVal"`   // 代理IP字符串 （包含）
	Username   string `json:"username"`   // 账号
	Password   string `json:"password"`   // 密码
	ExpireTime int64  `json:"expireTime"` // 过期时间
}

func (p *ProxyInfo) GetExpireTime() time.Time {
	return time.Unix(p.ExpireTime, 0).In(time.Local)
}

func (p *ProxyInfo) BuildProtocol() string {
	return fmt.Sprintf("%s://%s:%s@%s", p.Protocol, p.Username, p.Password, p.ProxyVal)
}

func (p *ProxyInfo) BuildChromeProtocol() string {
	return fmt.Sprintf("%s://%s", p.Protocol, p.ProxyVal)
}
