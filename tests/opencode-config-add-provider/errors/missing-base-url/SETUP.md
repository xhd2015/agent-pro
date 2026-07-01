# Scenario

**Feature**: missing --base-url is rejected

```
# --base-url omitted -> error mentions --base-url
agent-pro opencode config add-provider --id p --api-shape anthropic --model m1 -> error
```

## Preconditions

- All mandatory flags except `--base-url` are supplied.

## Steps

1. Set `req.Args` omitting `--base-url`.
2. Run and assert non-zero exit + stderr mentions `--base-url`; no config written.

## Context

- Isolates the `--base-url` required-flag validation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "p",
		"--api-shape", "anthropic",
		"--model", "m1",
	}
	return nil
}
```
