# Scenario

**Feature**: Remove when nothing installed only flushes sudo cache

```
# no drop-in/manifest -> only sudo -k, no error
```

## Preconditions

- No drop-in or manifest.

## Steps

1. Leave seeds empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "noop_when_missing"
	return nil
}
```