# Scenario

**Feature**: ExcludeSet with empty Exclude pattern is an error

```
FilterUserPrompts(ExcludeSet, Exclude="") -> error
```

## Preconditions

- Op filter; ExcludeSet=true; Exclude empty.

## Steps

1. Set empty exclude.
2. Call filter.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	req.ExcludeSet = true
	req.Exclude = ""
	return nil
}
```
