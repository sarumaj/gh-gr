package util

import (
	"os"
	"sync"

	logrus "github.com/sirupsen/logrus"
)

var logger = sync.Pool{New: initLogger}

func initLogger() any {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		FullTimestamp:          true,
	})
	return l
}

func FatalIfError(err error) {
	if err != nil {
		l := logger.Get().(*logrus.Logger)
		l.Fatalln(err)
	}
}

func Logger() *logrus.Logger {
	return logger.Get().(*logrus.Logger)
}
