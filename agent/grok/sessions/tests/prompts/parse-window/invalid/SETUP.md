# Scenario

**Feature**: invalid recent window tokens are rejected with clear errors

```
# empty | 0 unit | bare number | unknown unit
RecentRaw invalid -> error mentioning Nd, Nh, or Nm
```

## Preconditions

- Reject zero counts, empty string, bare numbers, and non-dhm units (e.g. `w`).

## Steps

1. Leaf sets invalid RecentRaw.
2. Assert error and guidance substrings.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```
