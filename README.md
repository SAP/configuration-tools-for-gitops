# The GitOps commandline interface

The GitOps CLI `coco` shall ultimately provide the central interface to interact
with the [gitops repository](https://github.tools.sap/MLF/mlf-gitops). It is
designed to provide extendable automations for the most common and/or cumbersome
tasks when dealing with the
[gitops repository](https://github.tools.sap/MLF/mlf-gitops).

## Code of conduct

- Above all, take code reviews seriously for this CLI.
  - Only accept what you understand (there is no shame in handing off reviews or
    asking for clarifications)
  - Only accept what is maintainable (sensible tests added, sensible package
    structure used, reasonable import of external packages)
- Eventually we want to open-source this CLI and the code quality should not
  stand in the way of internal or external contributers.
- Be direct but kind in discussions.

## Central CLI design descisions

`coco` prominently relies on the following open-source projects to provide
general functionalities:

- [cobra](https://github.com/spf13/cobra) for the CLI framework
- [viper](https://github.com/spf13/viper) for parameter, environment-variable
  and config-file consolidation
- [zap](https://github.com/uber-go/zap) for logging

## Package requirements

In order to have a well maintainable CLI the business logic for each command
must be implemented in its own package that is only loosly coupled to other
packages. In addition, tests for the business logic must be provided from day 1
and we aim for high test-coverage (80 % and upwords).

### Loose coupling

Strong coupling between packages can be broken up many times by consuming
interfaces instead of objects (`structs`) in all exported functions.

TODO add proper example

```go

```

### Idiomatic golang further reading

Here is an unstructured list of references aroung good code in golang and
idiomatic go:

- <https://talks.golang.org/2013/bestpractices.slide#17>
- <https://go.dev/doc/effective_go>
- <https://golangbyexample.com/interface-in-golang/>
- <https://dev.to/lcaparelli/should-my-methods-return-structs-or-interfaces-in-go-3b7>
-

## Entry points

- [getting-started](./docs/getting-started.md)
