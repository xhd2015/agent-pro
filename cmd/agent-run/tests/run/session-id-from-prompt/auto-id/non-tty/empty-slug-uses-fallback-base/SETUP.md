# Scenario

**Feature**: when slugify yields empty base, fall back to `sess` or `task`

```
agent-run run --session-id-from-prompt "!!!"
  -> base sess|task + -YYYYMMDD-HHMMSS
```

## Steps

1. Run with punctuation-only prompt `!!!`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "!!!"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
