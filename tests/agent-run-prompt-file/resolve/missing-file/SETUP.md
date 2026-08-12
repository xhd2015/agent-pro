# Scenario

**Feature**: missing `--prompt-file` path → error

```
ResolveRunPrompt("", /case/no-such-prompt.txt)
  -> error (does not exist / cannot read)
```

## Steps

1. Build absolute path under `d.DOCTEST_CASE` that is **not** created.
2. Positional empty; PromptFile = that path.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	path, err := missingCasePath(t, d, "no-such-prompt.txt")
	if err != nil {
		return err
	}
	req.Positional = ""
	req.PromptFile = path
	return nil
}
```
