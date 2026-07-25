# Scenario

**Feature**: FormatTree with Unicode connectors (default tree-cli style)

```
FormatTree(nodes, TreeFormatOptions{ASCII: false})
  -> lines use ├──  └──  │  (box-drawing)
```

## Preconditions

- `ASCII=false` (explicit).
- Same 3-node fixture as sibling ascii leaf.

## Steps

1. Set `req.ASCII = false`.
2. Assert TreeText matches the locked Unicode template (connectors + pids + cmds).

## Context

- Must **not** use ASCII-only connectors (`+--`, `` `-- ``, bare `|` as the
  sole pipe style without box-drawing) for the primary branch lines.
- Trailing newline required on the returned string (CLI-friendly).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ASCII = false
	return nil
}
```
