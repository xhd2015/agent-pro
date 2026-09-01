# Scenario

**Feature**: JSON format for grok models

```
# FormatJSON emits indented Catalog
Catalog -> FormatJSON -> {"home","default","models",...}
```

## Preconditions

- This branch tests `req.Format = "json"` only.

## Steps

1. Set `req.Format = "json"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Format = "json"
	return nil
}
```
