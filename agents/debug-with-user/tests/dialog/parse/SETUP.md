# Scenario

**Feature**: ParseOsascriptOutput extracts button and text answers

```
osascript stdout lines -> ParseOsascriptOutput -> button label, free text
```

## Preconditions

- Parser accepts typical `display alert` / `display dialog` stdout formats.

## Steps

1. Set `req.Operation = "parse"`.
2. Provide synthetic osascript stdout in `req.Input`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "parse"
	return nil
}
```
