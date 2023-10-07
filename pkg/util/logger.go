package util

import (
	"fmt"
	"os"
	"strings"

	logrus "github.com/sirupsen/logrus"
	tracerr "github.com/ztrue/tracerr"
)

var Logger = func() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	l.SetOutput(Stdout())
	l.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	return l
}()

func FatalIfError(err error) {
	if err == nil {
		return
	}

	err = tracerr.Wrap(err)

	var frames []string
	for _, frame := range err.(tracerr.Error).StackTrace() {
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

	Logger.SetOutput(Stderr())
	Logger.WithField("stack", frames).Fatalln(err)
}

func PrintlnAndExit(format string, a ...any) {
	n, _ := fmt.Fprintln(Stderr(), fmt.Sprintf(format, a...))
	os.Exit(max(n, 1))
}
