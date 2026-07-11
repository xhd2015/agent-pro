# Scenario

**Feature**: resume denied paths (exit 1)

```
not-exited | unbound | missing-session | missing-prompt -> exit 1 + clear error
```

## Steps

1. Leaf seeds the denying state and runs resume.
2. Assert exit 1 and error wording.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping marker for deny-path leaves (exit 1 expected).
	req.Mode = ""
	return nil
}
```
