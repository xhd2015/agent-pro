# Scenario

**Feature**: Active=true with empty active list yields empty result

```
sessions exist + active_sessions.sessions=[]
  -> empty list
```

## Preconditions

- At least one discoverable session.
- active_sessions.json present with empty sessions array.

## Steps

1. Write a session.
2. writeActiveSessions() with no ids.
3. Active=true, WantFormat=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.Active = true
	req.WantFormat = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "not active")
	writeActiveSessions(t, req.GrokHome /* none */)
	return nil
}
```
