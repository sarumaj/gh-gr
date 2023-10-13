package util

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	logrus "github.com/sirupsen/logrus"
	tracerr "github.com/ztrue/tracerr"
)

var Logger = func() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	c := Console()
	l.SetOutput(c.Stdout())
	l.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	return l
}()

func FatalIfError(err error, ignore ...error) {
	if err == nil {
		return
	}

	for _, e := range ignore {
		if errors.Is(err, e) {
			return
		}
	}

	err = tracerr.Wrap(err)

	var frames []string
	for _, frame := range err.(tracerr.Error).StackTrace() {
		switch ctx := frame.String(); {

		case
			strings.Contains(ctx, "FatalIfError()"),
			strings.Contains(ctx, "FatalIfErrorOrReturn[...]()"),
			strings.Contains(ctx, "runtime.main()"),
			strings.Contains(ctx, "runtime.goexit()"):

			continue

		default:
			frames = append(frames, frame.String())

		}
	}

	c := Console()
	Logger.SetOutput(c.Stderr())

	Logger.WithField("stack", frames).Fatalln(err)
}

func FatalIfErrorOrReturn[T any](arg T, err error) T {
	FatalIfError(err)
	return arg
}

func PrintlnAndExit(format string, a ...any) {
	c := Console()
	_ = FatalIfErrorOrReturn(fmt.Fprintln(c.Stderr(), c.CheckColors(color.RedString, format, a...)))
	os.Exit(1)
}
