# Scenario

**Feature**: Mode B lookup failures

```
fork.Main(["--session-id", unknown]) -> grok session not found
```

## Steps

1. Leaf uses an id that is not seeded.

## Context

- Error text matches `groksessions.Info`: `grok session not found`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--session-id", "019f283b-ffff-7fff-ffff-ffffffffffff"}
	return nil
}
```
