# Scenario

**Feature**: missing --api-shape is rejected

```
# --api-shape omitted -> error mentions --api-shape
agent-pro opencode config add-provider --id p --base-url ... --model m1 -> error
```

## Preconditions

- All mandatory flags except `--api-shape` are supplied.

## Steps

1. Set `req.Args` omitting `--api-shape`.
2. Run and assert non-zero exit + stderr mentions `--api-shape`; no config written.

## Context

- Isolates the `--api-shape` required-flag validation. The invalid-value case
  is covered by `invalid-api-shape`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "p",
		"--base-url", "https://api.example.com/v1",
		"--model", "m1",
	}
	return nil
}
```
