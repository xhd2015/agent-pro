# Scenario

**Feature**: `--prompt-file` is mutually exclusive with non-empty positional

```
prompt-file PATH (readable) + positional "x"
  -> ResolveRunPrompt -> error exclusive
```

## Steps

1. Write a readable prompt file under `d.DOCTEST_CASE`.
2. Set Positional=`x`, PromptFile to that path.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	path, err := writeCaseFile(t, d, "prompt.txt", fixturePromptFileRaw)
	if err != nil {
		return err
	}
	req.Positional = fixturePositional // "x"
	req.PromptFile = path
	return nil
}
```
