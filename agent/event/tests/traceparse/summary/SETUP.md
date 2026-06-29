# Scenario

**Feature**: summary helper functions

```
# mode selects which participant runs
doctest harness -> Run Mode=summary -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `summary`.

## Steps
1. Descendant leaves set `req.Mode` to `summary` (and wire-specific fields).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "summary"
	return nil
}
```
