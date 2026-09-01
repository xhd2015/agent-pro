# Scenario

**Feature**: human text format for codex models

```
# FormatText renders Home/Default and model lines with reasoning
Catalog -> FormatText -> human listing
```

## Preconditions

- This branch tests `req.Format = "human"` only.

## Steps

1. Set `req.Format = "human"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Format = "human"
	return nil
}
```
