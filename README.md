[![test_and_report](https://github.com/sarumaj/gh-gr/actions/workflows/test_and_report.yml/badge.svg)](https://github.com/sarumaj/gh-gr/actions/workflows/test_and_report.yml)
[![build_and_release](https://github.com/sarumaj/gh-gr/actions/workflows/build_and_release.yml/badge.svg)](https://github.com/sarumaj/gh-gr/actions/workflows/build_and_release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sarumaj/gh-gr)](https://goreportcard.com/report/github.com/sarumaj/gh-gr)
[![Maintainability](https://img.shields.io/codeclimate/maintainability-percentage/sarumaj/gh-gr.svg)](https://codeclimate.com/github/sarumaj/gh-gr/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/0c9cdd86241d58f97085/test_coverage)](https://codeclimate.com/github/sarumaj/gh-gr/test_coverage)
[![Go Reference](https://pkg.go.dev/badge/github.com/sarumaj/gh-gr/v2.svg)](https://pkg.go.dev/github.com/sarumaj/gh-gr/v2)
[![Go version](https://img.shields.io/github/go-mod/go-version/sarumaj/gh-gr?logo=go&label=&labelColor=gray)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/sarumaj/gh-gr?logo=github)](https://github.com/sarumaj/gh-gr/releases/latest)
[![Release Date](https://img.shields.io/github/release-date/sarumaj/gh-gr?logo=github)](https://github.com/sarumaj/gh-gr/releases/latest)
[![Commits since latest release](https://img.shields.io/github/commits-since/sarumaj/gh-gr/latest?logo=github)](https://github.com/sarumaj/gh-gr/releases/latest)
[![Downloads (all assets, all releases)](https://img.shields.io/github/downloads/sarumaj/gh-gr/total?logo=github)](https://github.com/sarumaj/gh-gr/releases)
[![Downloads (all assets, latest release)](https://img.shields.io/github/downloads/sarumaj/gh-gr/latest/total?logo=github)](https://github.com/sarumaj/gh-gr/releases/latest)

---

# gh-gr

**gh-gr** is a GitHub repository management tool based on the project [github-repo](https://github.com/CristianHenzel/github-repo) by [Cristian Henzel](https://github.com/CristianHenzel).

Since the original project used a configuration file containing sensitive information, the issue has been addressed by reinventing the tool as an extension to the [github cli (gh)](https://cli.github.com/).

[![gh-gr demo](doc/screenshot.png)](https://youtu.be/LLZ4bcqc_Bk)

## Installation

Prerequisites: [github cli (gh)](https://cli.github.com/)

To install gr:

```console
$ gh extension install https://github.com/sarumaj/gh-gr
```

## Usage

```console
$ gh gr --help

> gr is a gh cli extension allowing management of multiple repositories at once
>
> Usage:
>   gr [flags]
>   gr [command]
>
> Examples:
>   gh gr --concurrency 100 --timeout "20s" <subcommand>
>
> Available Commands:
>   cleanup     Clean up untracked local repositories
>   completion  Generate the autocompletion script for the specified shell
>   edit        Edit configuration
>   export      Export current configuration to stdout
>   help        Help about any command
>   import      Import configuration from stdin or a file
>   init        Initialize repository mirror
>   pull        Pull all repositories
>   push        Push all repositories
>   remove      Remove current configuration
>   status      Show status for all repositories
>   update      Update configuration
>   version     Display version information
>   view        Display current configuration
>
> Flags:
>   -c, --concurrency uint   Concurrency for concurrent jobs (default 12)
>   -h, --help               help for gr
>   -t, --timeout duration   Set timeout for long running jobs (default 10m0s)
>
> Use "gr [command] --help" for more information about a command.
```

First, create the configuration:

```console
$ gh gr init -d SOMEDIR -c 10
```

or, if you are willing to exclude some repositories, you can use regular expressions:

```console
$ gh gr init -c 10 -d SOMEDIR -e ".*repo1" -e "SOMEORG/repo-.*" -s
```

Run `gh gr init --help` or `gh gr help init` to retrieve more information about the init command.

After the configuration is created, you can pull all repositories using:

```console
$ gh gr pull
```

you can view the status of the repositories using:

```console
$ gh gr status
```

and you can push all repositories using:

```console
$ gh gr push
```

After creating new repositories on the server or after user data changes, you can update the local configuration using:

```console
$ gh gr update
```

## Acknowledgments

- [Cristian Henzel](https://github.com/CristianHenzel)
