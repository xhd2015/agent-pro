# Scenario

**Feature**: ParseRecentWindow accepts Nd|Nh|Nm only

```
# pure window parse — no filesystem
req.Op = "parse" ; RecentRaw = <token>
  -> ParseRecentWindow
  -> duration | error
```

## Preconditions

- No Grok session fixtures required (parse is pure).
- Valid form: one or more digits + unit `d`/`h`/`m` (case-insensitive).
- `1d` means 24 hours rolling (not calendar day).

## Steps

1. Set `Op=parse` and the raw window string.
2. Run ParseRecentWindow via harness.
3. Assert duration or error substrings.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "parse"
	return nil
}
```
