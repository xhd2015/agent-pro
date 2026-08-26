# Scenario

**Feature**: GrepSet with empty Grep pattern is an error

```
FilterUserPrompts(GrepSet, Grep="") -> error
```

## Preconditions

- Op filter; GrepSet=true; Grep empty string.

## Steps

1. Set empty grep.
2. Call filter.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	req.GrepSet = true
	req.Grep = []string{""}
	return nil
}
```
