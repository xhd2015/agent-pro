# Scenario

**Feature**: agent_trace backward compatibility

```
# mode selects which participant runs
doctest harness -> Run Mode=thin_wrapper -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `thin_wrapper`.

## Steps
1. Descendant leaves set `req.Mode` to `thin_wrapper` (and wire-specific fields).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "thin_wrapper"
	return nil
}
```
