# Scenario

**Feature**: weeks unit `"2w"` is rejected (only d/h/m)

```
ParseRecentWindow("2w") -> error
```

## Preconditions

- Input is `2w` (unsupported unit).

## Steps

1. Set RecentRaw to `2w`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "2w"
	return nil
}
```
