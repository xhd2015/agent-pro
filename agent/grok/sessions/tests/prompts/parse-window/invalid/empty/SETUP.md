# Scenario

**Feature**: empty recent window string is rejected

```
ParseRecentWindow("") -> error
```

## Preconditions

- Input is empty string.

## Steps

1. Set RecentRaw to empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentRaw = ""
	return nil
}
```
