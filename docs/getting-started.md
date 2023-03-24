# Getting started

## Local setup

To efficiently develope `coco`, the following tools must be installed in your
local environment:

- `make`
- `go` in version `go1.20` or higher
- `cobra` cli (run `go install github.com/spf13/cobra/cobra@latest`)
- for VSCode the `golang` extension

## Go run

For fast feedback you can use the `go run` command, e.g. as

```bash
go run cmd/coco/main.go --help
```

## Testing the CLI

The whole CLI can be tested (unit tests) by running

```bash
make test
```

In addition, individual tests can be run by using

```bash
test_function_name="..."
package_sub_folder="..."

TOP_LEVEL_DIR=$(git rev-parse --show-toplevel)
cd "${TOP_LEVEL_DIR}"

go test \
  -timeout 30s \
  -v \
  -race \
  -run "^${test_function_name}$" \
  "github.com/configuration-tools-for-gitops/${package_sub_folder}"
```

where you can get more info on `go test` by running `go help test`. In VSCode
with the golang plugin individual packages can be tested dynamically from the
IDE by opening the testfile and clicking on the dynamically appearing buttons
above the functions to test.

To obtain the coverage for a package you can run the following

```bash
package_sub_folder="..." # e.g. internal/dependencies

TOP_LEVEL_DIR=$(git rev-parse --show-toplevel)
cd "${TOP_LEVEL_DIR}"

go test \
  -timeout 30s \
  -v \
  -race \
  -coverprofile cover.out \
  "github.com/configuration-tools-for-gitops/${package_sub_folder}"
  go tool cover -func ./cover.out
```

### Running integration tests

To run integration tests, run

```bash
make integration
```

When adding a new package, make sure that the package is included at
[inttests.list](../integration-tests-list.txt)

## Building the CLI

### Local build

The CLI can be built locally to the target `dist/coco` by running

`make coco`

### Image build

For building the docker image there are the following option

```bash
# export DOCKER_PUSH=true ## to push the resulting docker image

## reproducible build (build context defined in build/Dockerfile)
make image

## developement build
export DEV_IMAGE=true
make image
```

## Folder structure

The entrypoint to `coco` is the top-level [main.go](./../main.go) file. As is
normal for a [cobra](https://github.com/spf13/cobra) application, this file
simply redirects you to the [rootCommand](./../cmd/root.go) in the
[cmd](./../cmd) package. This package holds the registrations for all commands
(including sub-commands).

### The `rootCommand`

The [rootCommand](./../cmd/root.go) represents the entrypoint into the command
hierarchy and additionally holds all logic concerning initializations such as
parsing the following inputs (in this order):

- configuration file if present (default location is `${HOME}/.coco.yaml`)
- environment variables
- command-line parameters

If variables are set multiple times their values will be overwritten by later
parsed values.

In addition the logging package is initialized.

### Subcommands

Subcommands can be registered by running

```bash
cobra add ${newCommand} -p ${parentCmd}
```

where `${parentCmd}` must be the corresponding variable name in the parent
command file. This will add a new file at `cmd/${newCommand}.go` that contains
the basic logic for commands. Add any additional commandline flags here.

Remove the auto-generated license section from the file `cmd/${newCommand}.go`
that has the form:

```go
/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
```

In the `Run` field of the command the handing off to the packages that implement
the business logic happens. The latter is located in different packages in the
`internal/` folder.

E.g. the `dependencies` command hands off to the
[dependencies package](./internal/dependencies/).
