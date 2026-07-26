# Scenario

**Feature**: empty `--resume-from-grok-session` value is rejected (parse or validation)

```
# empty flag value — no Grok fixture required
agent-run run --resume-from-grok-session=
  -> exit ≠ 0
```

## Steps

1. Run with equals-form empty id (always parseable as empty string by flags).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Empty value after equals form — always parseable as empty string by flags.
	req.Args = []string{"run", "--resume-from-grok-session="}
	return nil
}
```
