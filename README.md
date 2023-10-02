[![build_and_release](https://github.com/sarumaj/gh-gr/actions/workflows/build_and_release.yml/badge.svg)](https://github.com/sarumaj/gh-gr/actions/workflows/build_and_release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/sarumaj/gh-gr)](https://goreportcard.com/report/github.com/sarumaj/gh-gr)
[![Maintainability](https://img.shields.io/codeclimate/maintainability-percentage/sarumaj/gh-gr.svg)](https://codeclimate.com/github/sarumaj/gh-gr/maintainability)

---

**gh-gr** is a GitHub repository management tool based on the project [github-repo](https://github.com/CristianHenzel/github-repo) of [Cristian Henzel](https://github.com/CristianHenzel).

Since the original project used a configuration file containing sensitive information, the issue has been adressed by reinventing the tool as an extension to the [github cli (gh)](https://cli.github.com/).

# Installation

Prerequisites: [github cli (gh)](https://cli.github.com/)

To install gr:

```
gh extension install https://github.com/sarumaj/gh-gr
```

# Usage

First, create the configuration:

```
gh gr init -d SOMEDIR -c 10
```

or, if you are using GitHub Enterprise:

```
gh gr init -c 10 -r https://example.com/api/v3/ -d SOMEDIR -e "repo1|SOMEORG/repo-.*" -s
```

After the configuration is created, you can pull all repositories using:

```
gh gr pull
```

you can view the status of the repositories using:

```
gh gr status
```

and you can push all repositories using:

```
gh gr push
```

After creating new repositories on the server or after user data changes, you can update the local configuration using:

```
gh gr update
```

# Acknowledgments

- [Cristian Henzel](https://github.com/CristianHenzel)
