# Scenario

**Feature**: single PlaceCWD keeps only sessions under that abs cwd

```
sessions under cwdA and cwdB
  + PlaceCWDs=[abs(cwdA)]
  -> only cwdA session
```

## Preconditions

- Two sessions, different cwds.
- PlaceCWDs is a single absolute path for cwdA.

## Steps

1. Write idA1 under cwdA (newer) and idB1 under cwdB (older).
2. Set PlaceCWDs to absPath(cwdA), Limit=10.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "in place A")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "in place B")
	return nil
}
```
