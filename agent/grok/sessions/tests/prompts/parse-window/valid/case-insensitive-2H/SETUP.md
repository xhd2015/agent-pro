# Scenario

**Feature**: unit letter is case-insensitive (`2H` == `2h`)

```
ParseRecentWindow("2H") -> 2 * time.Hour
```

## Preconditions

- Input uses uppercase `H`.

## Steps

1. Set RecentRaw to `2H`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "2H"
	return nil
}
```
