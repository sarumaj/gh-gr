package config

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var logger = sync.Pool{New: initLogger}

func initLogger() any {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		FullTimestamp:          true,
	})
	return l
}

func Logger() *logrus.Logger {
	return logger.Get().(*logrus.Logger)
}
