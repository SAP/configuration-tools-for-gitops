# Configuration tools for GitOps
![Coverage](https://img.shields.io/badge/Coverage-85.4%25-brightgreen)

## About this project

CoCo (`configuration control` or officially known as
configuration-tools-for-gitops) is a command line interface designed to help
with configuration file management - especially for GitOps environments.

CoCo streamlines service configurations over many target environments by
offering:

- [file-generation](./cmd/coco/generate/readme.md):
  - global configuration alignment
  - exception marking in yaml configurations
- [dependency evaluation](./cmd/coco/dependencies/readme.md)
  - blast radius analysis of changes
- [dependency presentation](./cmd/coco/graph/readme.md)
  - structured representation of dependencies

The available commands of the CLI can be explored by running

```console
$ coco --help

coco is a CLI to interact with a gitops repository and shall provide
various solutions, ranging from file-generation over the calculation of
dependency trees to various interactions with git and github.

Usage:
  coco [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  dependencies Returns structured information which components and dependencies are affected by a change in git
  generate     generate allows to run file-generation over the gitops repository
  help         Help about any command
  version      coco version

Flags:
      --config string              config file (default $HOME/.coco)
  -b, --git-defaultbranch string   default branch (default "main")
      --git-depth int              [NOT IN USE (upstream bug: see https://github.com/go-git/go-git/issues/328 for issue tracking)]
                                                 max checkout depth of the git repository
  -p, --git-path string            path where the configuration repository locally resides
  -r, --git-remote string          remote branch to compare against for changed components (default "origin")
  -u, --git-url string             git URL of the configuration repository
  -h, --help                       help for coco
  -l, --loglvl level               sets the log level of the application - key or value of map[Debug:-1 Info:0 Warn:1 Error:2 DPanic:3 Panic:4 Fatal:5]

Use "coco [command] --help" for more information about a command.
```

## Requirements and Setup

To build CoCo locally, the following binaries must be available on your machine:

- [go](https://go.dev/doc/install)
- [make](https://www.gnu.org/software/make/)
- [yq](https://github.com/mikefarah/yq)
- [curl](https://curl.se/docs/manpage.html)
- [grep](https://www.gnu.org/software/grep/)
- [git](https://git-scm.com/)

The CoCo project can be built and tested using make. Please run `make help` to
see the available commands.

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via
[GitHub issues](https://github.com/SAP/configuration-tools-for-gitops/issues).
Contribution and feedback are encouraged and always welcome. For more
information about how to contribute, the project structure, as well as
additional contribution information, see our
[Contribution Guidelines](CONTRIBUTING.md).

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our
community a harassment-free experience for everyone. By participating in this
project, you agree to abide by its [Code of Conduct](CODE_OF_CONDUCT.md) at all
times.

## Licensing

Copyright 2023 SAP SE or an SAP affiliate company and
configuration-tools-for-gitops contributors. Please see our [LICENSE](LICENSE)
for copyright and license information. Detailed information including
third-party components and their licensing/copyright information is available
[via the REUSE tool](https://api.reuse.software/info/github.com/SAP/configuration-tools-for-gitops).
