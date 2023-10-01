package main

import (
	"github.com/sarumaj/gh-pr/pkg/commands"
)

// Version holds the application version.
// It gets filled automatically at build time.
var Version string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var BuildDate string

func main() {
	commands.Version = Version
	commands.BuildDate = BuildDate
	commands.Execute()
}
