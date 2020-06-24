# Contributing to ssm-helpers

:+1::tada: First off, thanks for taking the time to contribute! :tada::+1:

This document will help you understand the project more and how you can get involved.

The project is open sourced but it is mainly focused around use cases of making AWS Systems Manager Session Manager easier to use for developers.

Where possible we tried to keep similarities with `ssh` but we also wanted to make it easier for users to discover and connect to instances in AWS.
This especially includes managing instances based on instance tags in multiple accounts and in multiple regions.

The tools rely heavily on IAM permissions and should not require customization of the operating system.

To get started with AWS Session Manager please see their [getting started documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html).
We will not be able to provide support for AWS Session Manager or Systems Manager in your environment.
You should contact your AWS support staff for all issues and errors that are not directly related to these tools.

[Contributing](#contributing)

[Releasing](#releasing)

[Dependencies](#dependencies)

### Contributing

When you would like to contribute with code feel free to open an issue to explain the problem you're trying to solve.
We can help guide you in the right place to make that change.
In some cases improvements to these tools will need to be added to the [dependent libraries](#dependencies).

We welcome PRs for code changes but please keep in mind that the maintainers are volunteers and do not manage these repositories full time.

### Releasing

ssm-helpers uses tag based [semantic versioned](https://semver.org/) (semver) releases.
This means that if you push a tag named `v1.0.0` it will make a new [release](https://github.com/disneystreaming/ssm-helpers/releases) and build installation packages. It uses [Github Actions](https://github.com/features/actions) for CI and [goreleaser](https://goreleaser.com/) for CD.

### Dependencies

Each binary declares their own dependencies you can declare in the `main.go` file.
Dependencies are managed with `go mod` for local development and package releases.

Some special dependencies you may want to be aware of are:
  * [gomux](https://github.com/disneystreaming/gomux) library for managing tmux sessions with `ssm-session`
  * [go-ssm-helpers](https://github.com/disneystreaming/go-ssm-helpers) library for interacting with the AWS SSM API.
  * [aws-session-manager-plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html) The `brew install` method will install the AWS Session Manager plugin but that code is not open sourced and your computer will pull the releases directly from AWS. Updates to the plugin should PR the [homebrew-tap](https://github.com/disneystreaming/homebrew-tap/blob/master/Formula/aws-session-manager-plugin.rb#L4) repository.
