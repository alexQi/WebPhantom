package logger

import (
	"github.com/sirupsen/logrus"
)

type MineFormatter struct{}

func (f *MineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}
