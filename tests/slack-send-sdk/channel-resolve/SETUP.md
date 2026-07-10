# Scenario

**Feature**: channel name/ID resolution reflected in Sending line

```
slack-send [channel] -> resolveChannel -> print Sending to channel=RESOLVED -> send may fail (fake token)
doctest <- stdout Sending line with expected RESOLVED value
```

## Preconditions

- `valid-config.json` fixture with known `#general` → `C0ALE44K5J6`.
- Fake `botToken`; send fails after resolution (no `OK` line required).

## Steps

1. Load valid-config fixture into isolated workdir.
2. Leaf sets channel arg variant.
3. Assert stdout `Sending to channel=...` line; exit 1 with `send failed` in stderr.

## Context

- Tests pure resolution observable before/at send without requiring slacktest.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WriteGoMod = true
	req.UseRepoConfig = false
	req.ConfigFixture = "valid-config.json"
	req.SlackAPIURL = ""
	return nil
}
```