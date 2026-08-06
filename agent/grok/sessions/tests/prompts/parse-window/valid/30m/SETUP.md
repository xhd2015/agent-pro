# Scenario

**Feature**: `"30m"` parses to 30 minutes

```
ParseRecentWindow("30m") -> 30 * time.Minute
```

## Preconditions

- Input is exactly `30m`.

## Steps

1. Set RecentRaw to `30m`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "30m"
	return nil
}
```
