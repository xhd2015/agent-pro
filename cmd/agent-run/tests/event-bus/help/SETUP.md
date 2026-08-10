# Scenario

**Feature**: `run` help documents event-bus flags

```
agent-run run -h -> lists --event-bus-url and --event-bus-token
```

## Preconditions

- L2 pure: `agentruncli.RunHelpText()` returns the `run -h` help body
  (same text Handle prints via flags.Help).
- No binary build; no process stdio.

## Steps

1. Leaf sets `Op=help` (Args optional documentation of CLI shape).
2. `Run` calls `RunHelpText()`.
3. Assert documents both flags and trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opHelp
	return nil
}
```
