# Scenario

**Feature**: invalid filter option combinations surface clear errors

```
# FilterUserPrompts or ListPrompts validates opts before work
HeadSet+TailSet | Head=0 | Tail=0 | GrepSet+empty | ExcludeSet+empty
  -> error (non-nil); message mentions the problem
```

## Preconditions

- Prefer pure Op `filter` (no FS) so validation is isolated.
- Error text need not be byte-identical; ASSERT uses substring contracts.

## Steps

1. Set invalid combination on Request.
2. Call FilterUserPrompts via Op filter (empty input slice is fine).
3. Assert resp.Err is non-nil and contains expected substrings.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Op = "filter"
	req.FilterInput = nil
	return nil
}
```
