# Scenario

**Feature**: embedded newlines are escaped for AppleScript strings

```
multiline message -> Escape -> no raw newline characters in literal
```

## Steps

1. Use input spanning multiple lines (title + path block).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Input = "Step 1 — Did VS Code open?\nProject folder:\n/tmp/demo"
	return nil
}
```
