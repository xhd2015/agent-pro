# Scenario

**Feature**: FormatPromptsText / FormatPromptsListText compact CLI contract

```
# format surface
SessionPrompts | []SessionPrompts + FormatPromptsOptions{Location:UTC, Now}
  -> compact lines / headers / empty message
  -> trailing newline; no 👤 USER cards
```

## Preconditions

- Location forced to UTC for deterministic timestamps.
- Compact line layout: `[2006-01-02 15:04:05] text`
- Missing timestamp: `[—]`
- Multi header includes session id, relative last active, title, short cwd.

## Steps

1. Leaf prepares FS fixtures or synthetic SessionPrompts.
2. Calls format Op.
3. Assert text contract.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Location already UTC from root
	return nil
}
```
