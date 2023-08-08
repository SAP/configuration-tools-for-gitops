# Inputfile parsing

This package reads content of configuration files ('coco.yaml' if not otherwise specified) that are identified with a path.

Here is an example for inspecting the general structure and usage of config files
```bash
coco inspect values
```
Sample output: (Note: the output might differ due to modifications in the codebase)

```file
dependencies: list of dependencies ([]string)
name: name of component or environment (string) REQUIRED
type: type of the configuration file (string, options:[environment,component]) REQUIRED
values: list of .yaml files relative to the config file without file ending ([]string)
```

Both contents will be loaded in a `Coco` struct for the `Load` function to be independent of the content.
Unused values of the `Coco` struct will be nil.