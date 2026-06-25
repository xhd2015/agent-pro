# Scenario

**Feature**: legacy mode creates date-nested sess_* with questions, progress, and messages

```
# HOME/.agent-pro/subagent/layout-test/sessions/<date>/sess_*
subagent.Run -> events + messages + questions/ + progress/ in nested dir
```

## Preconditions

- Zero-value `SessionLayout`; no flat `Dir` pre-created.

## Steps

1. Inherit legacy base configuration.

## Context

- Inner session id from mock: `inner_legacy_sess`

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" {
		configureLegacyBase(t, req)
	}
	return nil
}```
