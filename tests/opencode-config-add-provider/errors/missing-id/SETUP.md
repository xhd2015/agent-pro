# Scenario

**Feature**: missing --id is rejected

```
# --id omitted -> error mentions --id
agent-pro opencode config add-provider --base-url ... --api-shape anthropic --model m1 -> error
```

## Preconditions

- All mandatory flags except `--id` are supplied.

## Steps

1. Set `req.Args` omitting `--id`.
2. Run and assert non-zero exit + stderr mentions `--id`; no config written.

## Context

- Isolates the `--id` required-flag validation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
	}
	return nil
}
```
