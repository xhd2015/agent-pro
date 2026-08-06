# Scenario

**Feature**: zero-count window `"0m"` is rejected

```
ParseRecentWindow("0m") -> error
```

## Preconditions

- Input is `0m` (zero is never a valid window).

## Steps

1. Set RecentRaw to `0m`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "0m"
	return nil
}
```
