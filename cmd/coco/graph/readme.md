# Graph plotting

The graph command presents a structured representation of the dependency graph
that is constructed in the [dependencies package](../dependencies/readme.md).

Possible output formats are `json`, `yaml`, and `flat`.

The following dependency graph

```test
A ← + ← C ← E
    ↑   ↑
    + ← D
```

would be encoded as

```yaml
A: {}
C:
  0: [A]
D:
  0: [A, C]
E:
  0: [C]
  1: [A]
```
