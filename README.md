# SSM Helpers

Helpers to manage you systems with [AWS Systems Manager](https://aws.amazon.com/systems-manager/) suite of management tools.

![](/img/ssm-helpers.gif)

## Tools in this repo

* `ssm` subcommands:
    
    * [`session`](cmd/ssm-session/README.md) - Interactive shell with an instance via AWS Systems Manager Session Manager (`ssh` and `cssh` replacement)

    * [`run`](cmd/ssm-run/README.md)     - Run a command on multiple instances based on instance tags or names (`mco` and `knife` replacement)

If you would like more information about the available commands, see the README for each in `./cmd/<command-name>/`.

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
# ssm run
docker run -it --rm -v $HOME/.aws:/root/.aws \
    -e AWS_PROFILE=$AWS_PROFILE -e AWS_REGION=$AWS_REGION \
    docker.pkg.github.com/disneystreaming/ssm-helpers/ssm run

# ssm session (change detach keys for tmux)
docker run -it --rm --detach-keys 'ctrl-e,e' \
    -v $HOME/.aws:/root/.aws -e AWS_PROFILE=$AWS_PROFILE \
    -e AWS_REGION=$AWS_REGION \
    docker.pkg.github.com/disneystreaming/ssm-helpers/ssm session
```

### Manually

You can find tagged releases for Windows, macOS, and Linux on the [releases page](https://github.com/disneystreaming/ssm-helpers)

### Latest from main

To install the package from git, you can fetch it via:

```
go get github.com/disneystreaming/ssm-helpers/
```

## Build

```
make build
```

### Testing

```
make test
```

### Linting

```
make check
```

## Develop

Each subcommand lives in the `cmd/<command-name>` folder and is written in go.  
They use the [aws-sdk-go](https://github.com/aws/aws-sdk-go) as well as our own fork of [gomux](https://github.com/disneystreaming/gomux).  
The [go-ssm-helpers](https://github.com/disneystreaming/go-ssm-helpers) library has been integrated into this project and will be archived.  
If you find bugs or would like to suggest improvements please use GitHub issues on the appropriate repo.
