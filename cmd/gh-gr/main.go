package main

import (
	commands "github.com/sarumaj/gh-gr/pkg/commands"
)

// Version holds the application version.
// It gets filled automatically at build time.
var Version = "v0.0.0"

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var BuildDate = "0000-00-00 00:00:00 UTC"

func main() {
	commands.Execute(Version, BuildDate)
}
