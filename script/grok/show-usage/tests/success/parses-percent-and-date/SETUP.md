# Scenario

**Feature**: parser extracts custom weekly limit and reset date from fake TUI

```
# fake TUI with non-default values -> stdout reflects parsed fields
grok-show-usage -> fake grok (42%, December 25, 12:00 UTC) -> print
```

## Preconditions

- Fake TUI returns different `Weekly limit` and `Next reset` values than the default fixture.

## Steps

1. Set `GROK_SHOW_USAGE_COMMAND` to `fakeTUICustom("42%", "December 25, 12:00 UTC")`.
2. Run and assert stdout matches the custom fixture.

## Context

- Verifies parsing is not hard-coded to the default `1%` / `July 9` values.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowUsageCommand = fakeTUICustom("42%", "December 25, 12:00 UTC")
	return nil
}
```