# Scenario

**Feature**: cmd/agent-run wires to agentruncli (thin main cutover)

```
# production sources
cmd/agent-run/*.go
  -> import "github.com/xhd2015/agent-pro/pkgs/agentruncli"
  -> agentruncli.Handle(...)
```

## Preconditions

- Scans production `.go` sources under `cmd/agent-run` (not `_test.go`).
- Does not require full command bodies moved yet for this leaf — import + Handle
  call prove the thin-main wire; behavior stays under existing CLI trees.

## Steps

1. Set harness mode `source_wire`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "source_wire"
	return nil
}
```
