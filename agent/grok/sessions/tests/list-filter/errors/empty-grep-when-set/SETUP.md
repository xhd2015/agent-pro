# Scenario

**Feature**: GrepSet with empty Grep returns an error

```
GrepSet=true, Grep=""
  -> error; does not treat as no filter
```

## Preconditions

- Mirrors prompts FilterUserPrompts empty-grep validation style.
- Fixtures optional.

## Steps

1. Set GrepSet=true, Grep="".
2. Limit=10.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.GrepSet = true
	req.Grep = ""
	req.Limit = 10
	return nil
}
```
