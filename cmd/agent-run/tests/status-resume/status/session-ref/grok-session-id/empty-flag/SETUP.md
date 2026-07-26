# Scenario

**Feature**: empty `--grok-session-id` value is rejected (parse or validation)

```
agent-run status --grok-session-id ""  (or --grok-session-id=)
  -> exit ≠ 0
```

## Steps

1. Run status with an empty grok-session-id value (no meta required).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Empty value after equals form — always parseable as empty string by flags.
	req.Args = []string{"status", "--grok-session-id="}
	return nil
}
```
