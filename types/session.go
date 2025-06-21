package types

import (
	"noctua/internal/model"
	"time"
)

// Session 代表账号会话，包括 Cookie 和代理信息
type Session struct {
	Enabled    bool
	Account    *model.MediaAccount // 账号信息
	ProxyInfo  *ProxyInfo          // 绑定的代理 IP
	ExpireTime time.Time           // 代理 IP 过期时间
	InUsed     bool
}
