# Scenario

**Feature**: `run` help documents idle-exit flags

```
agent-run run -h -> RunHelpText lists --exit-on-idle and --idle-timeout (default 10m)
```

## Preconditions

- L2 pure: `agentruncli.RunHelpText()` returns the `run -h` help body
  (same text Handle prints via flags.Help).
- No binary build; no process stdio.

## Steps

1. Grouping sets `Op=help`.
2. Leaf sets Args to `run -h` (documentary).
3. `Run` calls `RunHelpText()`.
4. Assert documents both flags, default `10m`, and trailing newline.

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
