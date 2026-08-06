# Scenario

**Feature**: HeadSet with Head=0 is an error (N >= 1 required)

```
FilterUserPrompts(HeadSet, Head=0) -> error
```

## Preconditions

- Op filter; HeadSet; Head=0.

## Steps

1. Set head zero.
2. Call filter.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	req.HeadSet = true
	req.Head = 0
	return nil
}
```
