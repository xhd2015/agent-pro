# Scenario

**Feature**: session ref without `--print` is rejected

```
seed session -> sessions web_test123 (no --print) -> exit 1
```

## Preconditions

- `--print` is required when a `<session_id>` positional is given.

## Steps

1. Seed any session so list-vs-print dispatch is unambiguous.
2. Run `agent-run sessions web_test123` without `--print`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printSessionID, "finished")
	req.Args = []string{"sessions", printSessionID}
	return nil
}
```
