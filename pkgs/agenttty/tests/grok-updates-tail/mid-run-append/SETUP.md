# Scenario

**Feature**: delayed appends while tail is watching must stream before ctx cancel

```
minimal seed on disk
  -> tail watches for new lines
  -> scheduled append delivers marker before cancel
```

## Preconditions

- Regression guard for WatchLine (or polling) delivery of post-bootstrap appends.
- Leaves schedule appends after `TailStartDelay`.

## Steps

1. Seed minimal initial content (often user chunk only).
2. Schedule tool/assistant append while tail is alive.
3. Assert marker in collected events.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartOffset = 0
	return nil
}
```