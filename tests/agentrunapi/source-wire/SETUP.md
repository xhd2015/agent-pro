# Scenario

**Feature**: cmd/agent-run wires to agentrunapi (CLI cutover exit criterion)

```
cmd/agent-run/*.go
  -> import "github.com/xhd2015/agent-pro/pkgs/agentrunapi"
```

## Preconditions

- Scans production `.go` sources under `cmd/agent-run` (not `_test.go`).
- Does not require implementing library body beyond import existence for this leaf;
  full behavior remains covered by auto-send-or-resume CLI integration tree.

## Steps

1. Set harness mode `source_wire`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	return nil
}
```
