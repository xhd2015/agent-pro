# Scenario

**Feature**: FormatTree with ASCII-only connectors

```
FormatTree(nodes, TreeFormatOptions{ASCII: true})
  -> lines use +--  `--  |  (no box-drawing)
```

## Preconditions

- `ASCII=true`.
- Same 3-node fixture as sibling unicode leaf.

## Steps

1. Set `req.ASCII = true`.
2. Assert TreeText matches the locked ASCII template.

## Context

- Must **not** emit Unicode box-drawing (`├`, `└`, `│`, `─`).
- Trailing newline required.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ASCII = true
	return nil
}
```
