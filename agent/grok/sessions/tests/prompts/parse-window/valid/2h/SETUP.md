# Scenario

**Feature**: `"2h"` parses to 2 hours

```
ParseRecentWindow("2h") -> 2 * time.Hour
```

## Preconditions

- Input is exactly `2h`.

## Steps

1. Set RecentRaw to `2h`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "2h"
	return nil
}
```
