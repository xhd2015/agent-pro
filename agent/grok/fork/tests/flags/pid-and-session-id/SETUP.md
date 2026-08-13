# Scenario

**Feature**: `--pid` and `--session-id` cannot be combined

```
fork.Main(["--pid", "6000", "--session-id", id]) -> error
```

## Steps

1. Args include both flags.

```go
import (
	"strconv"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--pid", strconv.Itoa(pidStart), "--session-id", fixtureSessionID}
	return nil
}
```
