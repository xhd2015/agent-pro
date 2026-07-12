# Scenario

**Feature**: `--new-terminal` without `--auto-send-or-resume` is rejected

```
agent-run run --new-terminal "hi"
  -> exit 1; error requires --auto-send-or-resume
```

## Steps

1. Run `run --new-terminal` with a prompt and no auto flag.
2. Clear iTerm capture path so a mistaken OpenConfig is not masked.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Validation only — do not enable iTerm hooks (feature must fail before open).
	req.ItermScriptOut = ""
	req.Args = []string{
		"run",
		"--new-terminal",
		"hi",
	}
	return nil
}
```
