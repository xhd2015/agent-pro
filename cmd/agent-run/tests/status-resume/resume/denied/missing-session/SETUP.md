# Scenario

**Feature**: resume missing session exits 1

```
agent-run resume no-such-resume-session "x" -> exit 1, not found
```

## Steps

1. Run resume against unknown id (no meta seed).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "no-such-resume-session"
	req.Args = []string{"resume", req.SessionID, "followup"}
	return nil
}
```
