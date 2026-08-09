# Scenario

**Feature**: whitespace-only forked_at is not treated as forked

```
session with forked_at="   \\t  " and empty kind
  + control fork session
  -> ListWithOptions(Forked=true)
  -> only kind=fork kept; whitespace forked_at excluded
```

## Preconditions

- forked_at must be non-empty **and** non-whitespace after trim.
- ForceWriteForkedAt so the key is present with whitespace value.

## Steps

1. Write idForkedAt with whitespace forked_at; idFork as positive control.
2. Forked=true.
3. Only idFork survives.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.Forked = true

	writeListSessionOpts(t, req.GrokHome, idForkedAt, atFixed(-20*time.Minute), cwdA, "whitespace forked_at", listSessionOpts{
		ForkedAt:           "  \t  ",
		ForceWriteForkedAt: true,
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-10*time.Minute), cwdA, "real fork", listSessionOpts{
		SessionKind: "fork",
	})
	return nil
}
```
