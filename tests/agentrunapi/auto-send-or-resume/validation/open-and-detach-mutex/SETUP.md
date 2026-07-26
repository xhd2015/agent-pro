# Scenario

**Feature**: Open and Detach are mutually exclusive

```
AutoSendOrResume(Opts{SessionID, Open:true, Detach:true})
  -> error (mutual exclusive)
  -> zero hooks
```

## Preconditions

- Non-empty SessionID so open/detach mutex is the failing gate (not empty id).
- Hooks installed; must not fire.

## Steps

1. Set Open and Detach both true with valid SessionID.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-mutex-1"
	req.Open = true
	req.Detach = true
	req.SeedMeta = false
	return nil
}
```
