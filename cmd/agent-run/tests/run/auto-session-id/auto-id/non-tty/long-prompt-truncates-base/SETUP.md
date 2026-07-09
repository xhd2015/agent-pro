# Scenario

**Feature**: base slug is truncated to ≤ 128 runes before timestamp is appended

```
agent-run run --auto-session-id <200 x 'a'>
  -> base length ≤ 128 runes
  -> timestamp still appended after base
```

## Steps

1. Build a prompt of 200 ASCII `a` characters (no separators).
2. Run with `--auto-session-id`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Prompt = strings.Repeat("a", 200)
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
