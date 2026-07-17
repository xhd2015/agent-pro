# Scenario

**Feature**: empty/whitespace SessionID rejected without dispatch

```
AutoSendOrResume(Opts{SessionID: "  "})
  -> error mentions session
  -> zero hooks
```

## Preconditions

- SessionID blank after trim.
- Hooks installed by parent; must not fire.

## Steps

1. Set SessionID to whitespace-only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "   \t  "
	req.SeedMeta = false
	return nil
}
```
