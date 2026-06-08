# 3-depth 24-case example

## DOT graph
```mermaid
graph TD
  root[Project permissions] --> owner[owner flows]
  root --> maintainer[maintainer flows]
  root --> viewer[viewer flows]
  owner --> ownerCases[8 owner test leaves]
  maintainer --> maintainerCases[8 maintainer test leaves]
  viewer --> viewerCases[8 viewer test leaves]
```

## Text tree
- Project permissions
  - owner: 8 runnable leaves
  - maintainer: 8 runnable leaves
  - viewer: 8 runnable leaves

## File index
This example demonstrates a root, a mode layer, and leaf test cases at depth 3.
