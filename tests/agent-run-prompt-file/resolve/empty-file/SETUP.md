# Scenario

**Feature**: empty `--prompt-file` yields empty prompt (after trim)

```
empty file + no positional
  -> ResolveRunPrompt -> ""
```

## Steps

1. Write zero-byte `empty.txt` under `d.DOCTEST_CASE`.
2. Positional empty; PromptFile = that path.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	path, err := writeCaseFile(t, d, "empty.txt", "")
	if err != nil {
		return err
	}
	req.Positional = ""
	req.PromptFile = path
	return nil
}
```
