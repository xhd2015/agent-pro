# Scenario

**Feature**: Escape produces AppleScript-safe string literals

```
raw title/message with metacharacters -> Escape -> doubled quotes/backslashes, escaped newlines
```

## Preconditions

- `dialog.Escape` is exported and deterministic.

## Steps

1. Set `req.Operation = "escape"`.
2. Provide adversarial `req.Input` with quotes, backslashes, or newlines.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "escape"
	return nil
}
```
