# SSM Helpers

Helpers to manage you systems with [AWS Systems Manager](https://aws.amazon.com/systems-manager/) suite of management tools.

![](/img/ssm-helpers.gif)

## Tools in this repo

* [`ssm-session`](cmd/ssm-session/README.md) - Interactive shell with an instance via AWS Systems Manager Session Manager (`ssh` and `cssh` replacement)
* [`ssm-run`](cmd/ssm-run/README.md)     - Run a command on multiple instances based on instance tags or names (`mco` and `knife` replacement)

If you would like more information about the tools, see the README for each in `./cmd/<tool-name>/`.

## Install

![goreleaser](https://github.com/disneystreaming/ssm-helpers/workflows/goreleaser/badge.svg)

### Homebrew

Install the tools via homebrew with

```
brew tap disneystreaming/ssm-helpers

brew install ssm-helpers
```

For more information on Homebrew taps please see the [tap documentation](https://docs.brew.sh/Taps)

### Manually

You can find tagged releases for Windows, macOS, and Linux on the [releases page](https://github.com/disneystreaming/ssm-helpers)

### Latest from master

To install the packages from git you can install them individually via:

```
go get github.com/disneystreaming/cmd/ssm-session

go get github.com/disneystreaming/cmd/ssm-run
```

## Build

```
go build -o ssm-run ./cmd/ssm-run/main.go
go build -o ssm-session ./cmd/ssm-session/main.go
```
