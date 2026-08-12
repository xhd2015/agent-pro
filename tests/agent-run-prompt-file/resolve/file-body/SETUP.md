# Scenario

**Feature**: `--prompt-file` body becomes the resolved prompt (TrimSpace)

```
file body "  hello\n" + empty positional
  -> ResolveRunPrompt -> "hello"
```

## Steps

1. Write `prompt.txt` under `d.DOCTEST_CASE` with raw body `fixturePromptFileRaw`.
2. Set Positional empty, PromptFile to that absolute path.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	path, err := writeCaseFile(t, d, "prompt.txt", fixturePromptFileRaw)
	if err != nil {
		return err
	}
	req.Positional = ""
	req.PromptFile = path
	return nil
}
```
