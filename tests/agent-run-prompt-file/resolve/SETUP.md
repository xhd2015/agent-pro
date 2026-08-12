# Scenario

**Feature**: pure `ResolveRunPrompt` loads file / enforces exclusive / errors

```
ResolveRunPrompt(positional, promptFile)
  -> prompt | error
```

## Steps

1. Set mode `resolve`.
2. Leaf writes fixtures under `d.DOCTEST_CASE` and sets Positional / PromptFile.
3. Assert prompt string or error.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "resolve"
	return nil
}
```
