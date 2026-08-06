# Scenario

**Feature**: bare number without unit is rejected

```
ParseRecentWindow("30") -> error
```

## Preconditions

- Input is `30` with no `d`/`h`/`m` suffix.

## Steps

1. Set RecentRaw to `30`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = "30"
	return nil
}
```
