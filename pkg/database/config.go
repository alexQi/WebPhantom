package database

// Config 数据库配置结构体
type Config struct {
	Driver   string // 数据库类型: mysql, postgres, sqlite
	DSN      string // 连接字符串
	LogLevel int    // 日志级别: silent, error, warn, info
}
