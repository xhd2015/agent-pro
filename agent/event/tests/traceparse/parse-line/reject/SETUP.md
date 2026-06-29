# Scenario

**Feature**: lines that must not parse

```
# mode selects which participant runs
doctest harness -> Run Mode=parse_line -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `parse_line`.

## Steps
1. Descendant leaves set `req.Mode` to `parse_line` (and wire-specific fields).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "parse_line"
	return nil
}
```
