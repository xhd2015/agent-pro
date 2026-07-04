# Scenario

**Feature**: missing codex binary without fake hook fails clearly

```
# no codex on PATH, no CODEX_SHOW_STATUS_COMMAND -> error mentions codex
codex-show-status -> resolve codex path -> not found
```

## Preconditions

- `PATH` contains no `codex` executable.
- `CODEX_SHOW_STATUS_COMMAND` is not set.

## Steps

1. Set `req.SkipFakeCommand = true` and `req.MinimalPATH = true`.
2. Run and assert non-zero exit; stderr mentions `codex`.

## Context

- Exercises production argv resolution when codex is absent.
- `tty-watch` remains on PATH via minimal PATH construction in `Run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SkipFakeCommand = true
	req.MinimalPATH = true
	return nil
}
```