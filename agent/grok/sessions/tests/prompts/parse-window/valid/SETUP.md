# Scenario

**Feature**: valid recent window tokens parse to positive durations

```
# valid Nd|Nh|Nm
RecentRaw valid -> duration > 0, no error
```

## Preconditions

- Token matches `^([0-9]+)([dhm])$` with non-zero count.

## Steps

1. Leaf sets RecentRaw.
2. Assert Window equals expected duration.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Op already parse from parent
	return nil
}
```
