# Scenario

**Feature**: `--session-id-from-prompt` generates id from prompt slug

```
agent-run run --session-id-from-prompt "prompt"
  -> id = slugify(prompt)-YYYYMMDD-HHMMSS[-N]
  -> storage sessions/<runner>/<id>/
```

## Preconditions

- `--session` is not set.
- Prompt is non-empty after trim.

## Steps

1. Grouping adds `--session-id-from-prompt` to `run` args.
2. Child dirs split non-TTY vs TTY runners and slug edges.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "--session-id-from-prompt")
	return nil
}
```
