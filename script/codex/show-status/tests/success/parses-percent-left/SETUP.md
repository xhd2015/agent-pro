# Scenario

**Feature**: parser converts percent-left to monthly usage rate

```
# fake TUI with 30% left -> stdout Monthly usage: 70%
codex-show-status -> fake codex (30% left) -> print
```

## Preconditions

- Fake TUI returns `30% left` on the monthly credit limit line (not the default `42% left`).

## Steps

1. Set `CODEX_SHOW_STATUS_COMMAND` to `fakeTUICustom("30%", "3000", "10,000", "12:00 on 15 Jan")`.
2. Run and assert stdout reflects `70%` monthly usage and matching credits/reset.

## Context

- Verifies `monthly_usage = (100 - N)%` is not hard-coded to the default `58%` / `42% left` values.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowStatusCommand = fakeTUICustom("30%", "3000", "10,000", "12:00 on 15 Jan")
	return nil
}
```