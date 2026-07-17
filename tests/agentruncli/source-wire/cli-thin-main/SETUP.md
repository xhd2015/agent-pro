# Scenario

**Feature**: agent-run main package imports agentruncli and calls Handle

```
# thin wrapper contract
parse cmd/agent-run/*.go
  -> pkgs/agentruncli import present
  -> Handle call present
```

## Preconditions

- Module layout: DOCTEST_ROOT = tests/agentruncli → `../../cmd/agent-run`.

## Steps

1. Mode already `source_wire` from parent; reinforce.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "source_wire"
	return nil
}
```
