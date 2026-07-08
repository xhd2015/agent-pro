# Scenario

**Bug**: consumer `WatchEvents` / CLI follow must not gate on session status

```
finished session + live watch ctx
  -> append to events.jsonl
  -> consumer delivers new line before ctx ends
```

## Steps

1. Grouping setup sets `req.Mode = "cli-follow"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "cli-follow"
	return nil
}
```