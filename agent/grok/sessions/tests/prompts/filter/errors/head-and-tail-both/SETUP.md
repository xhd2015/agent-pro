# Scenario

**Feature**: HeadSet and TailSet together is an error

```
FilterUserPrompts(HeadSet+TailSet) -> error
```

## Preconditions

- Op filter; HeadSet Head=1; TailSet Tail=1.
- FilterInput may be empty/nil.

## Steps

1. Set both head and tail.
2. Call filter.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	req.HeadSet = true
	req.Head = 1
	req.TailSet = true
	req.Tail = 1
	return nil
}
```
