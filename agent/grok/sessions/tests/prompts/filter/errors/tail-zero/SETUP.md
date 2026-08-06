# Scenario

**Feature**: TailSet with Tail=0 is an error (N >= 1 required)

```
FilterUserPrompts(TailSet, Tail=0) -> error
```

## Preconditions

- Op filter; TailSet; Tail=0.

## Steps

1. Set tail zero.
2. Call filter.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	req.TailSet = true
	req.Tail = 0
	return nil
}
```
