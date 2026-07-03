# Scenario

**Feature**: quotes and backslashes are doubled for AppleScript

```
input with " and \ -> Escape -> \" and \\
```

## Steps

1. Use input containing both double-quote and backslash characters.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Input = `say "hi"\path`
	return nil
}
```
