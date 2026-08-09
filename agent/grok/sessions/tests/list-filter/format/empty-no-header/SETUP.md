# Scenario

**Feature**: empty FormatListTable stays "No sessions found" (no KIND header)

```
empty sessions tree
  -> ListWithOptions -> []
  -> FormatListTable -> "No sessions found"
```

## Preconditions

- WantFormat=true.
- Empty discover result; phrase unchanged by KIND feature.

## Steps

1. Leave sessions empty.
2. WantFormat=true.
3. Assert exact phrase.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.WantFormat = true
	return nil
}
```
