package util

import (
	"os"
	"strings"

	logrus "github.com/sirupsen/logrus"
	"github.com/ztrue/tracerr"
)

var Logger = func() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
	return l
}()

func FatalIfError(err error) {
	if err != nil {
		err := tracerr.Wrap(err)

		var frames []string
		for _, frame := range err.StackTrace() {
			switch ctx := frame.String(); {

			case
				strings.Contains(ctx, "FatalIfError()"),
				strings.Contains(ctx, "runtime.main()"),
				strings.Contains(ctx, "runtime.goexit()"):

				continue

			default:
				frames = append(frames, frame.String())

			}
		}

		Logger.SetOutput(os.Stderr)
		Logger.WithField("stack", frames).Fatalln(err)
	}
}
