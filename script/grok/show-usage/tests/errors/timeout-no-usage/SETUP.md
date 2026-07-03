# Scenario

**Feature**: timeout when fake TUI never prints usage lines

```
# fake TUI hangs after prompt -> wait exceeds GROK_SHOW_USAGE_TIMEOUT -> timeout error
grok-show-usage -> fake grok (no Weekly limit) -> timeout
```

## Preconditions

- Fake TUI prints prompt and reads `/usage show` but never emits `Weekly limit:`.
- `GROK_SHOW_USAGE_TIMEOUT` shortened to 3 seconds for fast test.

## Steps

1. Set `ShowUsageCommand` to `fakeTUINoUsage()`.
2. Set `TimeoutSeconds` to `"3"`.
3. Assert non-zero exit; stderr mentions `timeout`.

## Context

- Default CLI timeout is 30s; this leaf overrides to keep CI fast.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowUsageCommand = fakeTUINoUsage()
	req.TimeoutSeconds = "3"
	return nil
}
```