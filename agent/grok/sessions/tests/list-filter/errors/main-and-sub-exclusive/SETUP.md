# Scenario

**Feature**: MainAgent and SubAgent together return an error

```
MainAgent=true + SubAgent=true
  -> ListWithOptions error (mutually exclusive); no panic
```

## Preconditions

- Mutual exclusion enforced by library (not only CLI).
- Fixtures optional; validation should not require sessions.

## Steps

1. Set MainAgent and SubAgent true.
2. Leave Forked false.
3. Assert resp.Err mentions main/sub (or mutually exclusive / exclusive).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.MainAgent = true
	req.SubAgent = true
	return nil
}
```
