# Scenario

**Feature**: `"1d"` parses to 24 hours (rolling day)

```
ParseRecentWindow("1d") -> 24 * time.Hour
```

## Preconditions

- Input is exactly `1d`.
- Day unit is 24h rolling, not calendar midnight.

## Steps

1. Set RecentRaw to `1d`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "1d"
	return nil
}
```
