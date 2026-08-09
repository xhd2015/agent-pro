# Scenario

**Feature**: place AND MainAgent still AND (role after place in pipeline)

```
# main in A kept
# sub in A dropped (wrong role)
# main in B dropped (wrong place)
PlaceCWDs=[A] + MainAgent=true
```

## Preconditions

- Place filter OR within PlaceCWDs; role ANDs with place.
- One leaf proving role does not break place pipeline.

## Steps

1. Write main-in-A, sub-in-A, main-in-B.
2. PlaceCWDs=[A], MainAgent=true.
3. Only main-in-A remains.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	req.MainAgent = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-30*time.Minute), cwdA, "main in A keep", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-20*time.Minute), cwdA, "sub in A drop", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idEmptyNo, atFixed(-10*time.Minute), cwdB, "main in B drop", listSessionOpts{})
	return nil
}
```
