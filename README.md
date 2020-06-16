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
brew install disneystreaming/tap/ssm-helpers
```

For more information on Homebrew taps please see the [tap documentation](https://docs.brew.sh/Taps)

### Docker

You can run the tools from docker containers

```bash
# ssm-run
docker run -it --rm -v $HOME/.aws:/root/.aws \
    -e AWS_PROFILE=$AWS_PROFILE -e AWS_REGION=$AWS_REGION \
    docker.pkg.github.com/disneystreaming/ssm-helpers/ssm-run

# ssm-session (change detach keys for tmux)
docker run -it --rm --detach-keys 'ctrl-e,e' \
    -v $HOME/.aws:/root/.aws -e AWS_PROFILE=$AWS_PROFILE \
    -e AWS_REGION=$AWS_REGION \
    docker.pkg.github.com/disneystreaming/ssm-helpers/ssm-session
```

### Manually

You can find tagged releases for Windows, macOS, and Linux on the [releases page](https://github.com/disneystreaming/ssm-helpers)

### Latest from master

To install the packages from git you can install them individually via:

```
go get github.com/disneystreaming/ssm-helpers/cmd/ssm-session

go get github.com/disneystreaming/ssm-helpers/cmd/ssm-run
```

## Build

```
go build -o ssm-run ./cmd/ssm-run/main.go
go build -o ssm-session ./cmd/ssm-session/main.go
```

## Develop

Each tool lives in the `cmd/<tool-name>` folder and is written in go.
They use the [aws-sdk-go](https://github.com/aws/aws-sdk-go) as well as our own fork of [gomux](https://github.com/disneystreaming/gomux) and a new library we provide at [go-ssm-helpers](https://github.com/disneystreaming/go-ssm-helpers).

If you find bugs or would like to suggest improvements please use GitHub issues on the appropriate repo.
