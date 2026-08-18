# Scenario

**Feature**: empty / whitespace query id rejected by Find and List alike

```
FindByGrokSessionID("") / ("   ")
ListByRunnerSessionID("") / ("   ")
  -> error: --grok-session-id requires a non-empty value
```

## Steps

1. No seeds required.
2. Op `find_and_list` with whitespace query id (covers empty-after-trim).
3. Assert both Find and List errors are the exact empty-id message.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "find_and_list"
	req.QueryID = "   \t  "
	req.Runners = []string{"grok", "grok-tty"}
	return nil
}
```
