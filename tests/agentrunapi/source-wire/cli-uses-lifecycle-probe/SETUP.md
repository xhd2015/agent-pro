# Scenario

**Feature**: agentruncli production path wires LifecycleProbe

```
pkgs/agentruncli/*.go (run_cmd / status)
  -> agentrunapi.LifecycleProbe referenced
```

## Preconditions

- Module layout: d.DOCTEST_ROOT = tests/agentrunapi → `../../pkgs/agentruncli`.
- Production non-test sources only.

## Steps

1. Mode `source_wire`; SourceWireTarget `agentruncli`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	req.SourceWireTarget = "agentruncli"
	return nil
}
```
