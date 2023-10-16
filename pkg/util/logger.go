package util

import (
	"fmt"
	"os"
	"strings"

	color "github.com/fatih/color"
	supererrors "github.com/sarumaj/go-super/errors"
	logrus "github.com/sirupsen/logrus"
	tracerr "github.com/ztrue/tracerr"
)

// App logger (default format JSON).
var Logger = func() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.WarnLevel)
	c := Console()
	l.SetOutput(c.Stdout())
	l.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	supererrors.RegisterCallback(func(err error) {
		c := Console()
		l.SetOutput(c.Stderr())

		if l.Level >= logrus.DebugLevel {
			err = tracerr.Wrap(err)

			var frames []string
			for _, frame := range err.(tracerr.Error).StackTrace() {
				switch ctx := frame.String(); {

				case
					strings.Contains(ctx, "supererrors.Except"),
					strings.Contains(ctx, "runtime.main()"),
					strings.Contains(ctx, "runtime.goexit()"):

					continue

				default:
					frames = append(frames, frame.String())

				}
			}

			l.WithField("stack", frames).Fatalln(err)
		}

		l.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:       true,
			DisableLevelTruncation: true,
		})
		l.Fatalln(err)
	})

	return l
}()

// Print to Stdout and exit.
func PrintlnAndExit(format string, a ...any) {
	c := Console()
	_ = supererrors.ExceptFn(supererrors.W(fmt.Fprintln(c.Stderr(), c.CheckColors(color.RedString, format, a...))))
	os.Exit(1)
}
