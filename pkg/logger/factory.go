package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

type LoggerConfig struct {
	Stdout bool
	Level  string
	Path   string
}

var Log *logrus.Logger

func Init(cfg *LoggerConfig) {
	var logLevel logrus.Level
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		DisableColors:   false,
		FullTimestamp:   true,
		DisableQuote:    true,
		TimestampFormat: "15:04:05",
	})

	if len(cfg.Level) == 0 {
		cfg.Level = "info"
	}

	err := logLevel.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		logLevel = 4
		Log.Infof("Use default log Level: INFO")
	}
	Log.SetLevel(logLevel)
	if cfg.Stdout {
		Log.Out = os.Stdout
	}
	if len(cfg.Path) > 0 {
		_, err = os.Stat(cfg.Path)
		if err != nil {
			err = os.MkdirAll(cfg.Path, os.ModePerm)
			if err != nil {
				Log.Panicf("mkdir error : %s", err.Error())
			}
		}
		NewSimpleLogger(Log, cfg.Path, 30, &logrus.TextFormatter{
			ForceColors:     true,
			DisableColors:   false,
			FullTimestamp:   true,
			DisableQuote:    true,
			TimestampFormat: "15:04:05",
		})
	}
	Log.Infof("Log successfully inited, path %s", cfg.Path)
}

/*
*

	文件日志
*/
func NewSimpleLogger(log *logrus.Logger, logPath string, save uint, formatter logrus.Formatter) {

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer(logPath, "debug", save),
		logrus.TraceLevel: writer(logPath, "trace", save),
		logrus.InfoLevel:  writer(logPath, "info", save),
		logrus.WarnLevel:  writer(logPath, "warn", save),
		logrus.ErrorLevel: writer(logPath, "error", save),
		logrus.FatalLevel: writer(logPath, "fatal", save),
		logrus.PanicLevel: writer(logPath, "panic", save),
	}, formatter)
	log.AddHook(lfHook)
}

func writer(logPath string, level string, save uint) *rotatelogs.RotateLogs {
	logFullPath := path.Join(logPath, level)
	logFullPath = fmt.Sprintf("%s", logFullPath)

	logier, err := rotatelogs.New(
		logFullPath+"-%Y%m%d.log",
		rotatelogs.WithRotationTime(time.Second), // 日志切割时间间隔
		rotatelogs.WithMaxAge(-1),                // 关闭过期清理
		rotatelogs.WithRotationCount(int(save)),  // 文件最大保存份数
		//rotatelogs.WithLinkName(logFullPath+".out"), // 生成软链，指向最新日志文件
	)

	if err != nil {
		panic(err)
	}
	// **在应用启动时，确保 `log_lock` 被清理**
	lockFile := logFullPath + "-" + time.Now().Format("20060102") + ".log_lock"
	if _, err := os.Stat(lockFile); err == nil {
		_ = os.Remove(lockFile)
	}
	return logier
}
