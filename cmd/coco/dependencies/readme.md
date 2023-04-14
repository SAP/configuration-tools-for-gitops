# Dependency naming

This package constructs the directed dependency graph for all components under
the `git.path`. A component is identified by a `coco.yaml` file with the format:

```yaml
# coco.yaml content

name: name of the component
dependencies:
  - list of
  - all other components
  - that this component
  - depends on directly
```

In the `coco.yaml` files dependencies are given from `Downstream` to `Upstream`,
meaning that this file contains all dependencies that this component requires to
work (a.k.a `Upstream` components). Therefore, the directed graph that can be
constructed from the `coco.yaml` files flows from `Downstream` to `Upstream`
components. For the identification which `Downstream` components are potentially
affected by an `Upstream` change (and hence need to be tested in change
validation), this flow direction is the inverse of what we need.

In this package the graph construction from a set of `coco.yaml` files and the
inversion of the graph is done. For later consumption the inversed graph is
represented from each node as map of maps. See the following examples

```sh
# service A depends on upstream services C and D

# graph from Downstream to Upstream

A -+ → C
   |
   + → D

# graph from Upstream to Downstream

C → A

D → A

# desired output

map[
  A: [],
  C: map[0:map[A:true]],
  D: map[0:map[A:true]],
]
```

```sh
# service A depends on upstream services C and D

# graph from Downstream to Upstream

A → + → C → E
    ↓   ↓
    + → D

# graph from Upstream to Downstream

A ← + ← C ← E
    ↑   ↑
    + ← D

# desired output

map[
  A: map[],
  C: map[0:map[A:true]],
  D: map[
    0:map[
      A:true,
      C:true,
    ]
  ],
  E: map[
    0:map[C:true],
    1:map[A:true],
  ]
]
```

<pre>
arrows
&#8595 ↓
&#8593 ↑
&#8594 →
&#8592 ←
&#8627 ↳
&#8625 ↱
</pre>
