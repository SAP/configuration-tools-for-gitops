# Inputfile parsing

This package deals with configuration files ('coco.yaml' if not otherwise
specified) that are identified within the basepath.

All possible inputs for configuration files can be obtained by running

```bash
coco inspect values
```

Sample output: (Note: the output might differ due to modifications in the
codebase)

```file
dependencies: list of dependencies ([]string)
name: name of component or environment (string) REQUIRED
type: type of the configuration file (string, options:[environment,component]) REQUIRED
values: list of .yaml files relative to the config file without file ending ([]string)
```
