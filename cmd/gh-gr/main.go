/*
	gh-gr is a CLI tool to manage GitHub repositories.
	The tool is designed to work with multiple GitHub accounts and organizations.
  It utilizes the GitHub API to retrieve the list of repositories and their metadata.
  Futhermore, it plugs into the GitHub CLI to provide a seamless experience.

	Usage:
    gr [flags]
    gr [command]

  Available Commands:
    cleanup     Clean up untracked local repositories
    completion  Generate the autocompletion script for the specified shell
    export      Export current configuration to stdout
    help        Help about any command
    import      Import configuration from stdin or a file
    init        Initialize repository mirror
    pull        Pull all repositories
    push        Push all repositories
    remove      Remove current configuration
    status      Show status for all repositories
    update      Update configuration
    version     Display version information
    view        Display current configuration

  Flags:
    -c, --concurrency uint   Concurrency for concurrent jobs (default 12)
    -h, --help               help for gr
    -t, --timeout duration   Set timeout for long running jobs (default 10m0s)

  Use "gr [command] --help" for more information about a command.
*/

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
