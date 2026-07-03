# Scenario

**Feature**: parse failure when TUI output lacks usage fields

```
# fake TUI prints garbage after /usage show -> parse error
grok-show-usage -> fake grok (no Weekly limit/Next reset) -> parse failure
```

## Preconditions

- Fake TUI responds to `/usage show` but prints unparseable output.

## Steps

1. Set `ShowUsageCommand` to `fakeTUIMalformed()`.
2. Assert non-zero exit; stderr mentions `parse`.

## Context

- Distinguishes parse errors from timeout (TUI responds quickly with bad data).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowUsageCommand = fakeTUIMalformed()
	return nil
}
```