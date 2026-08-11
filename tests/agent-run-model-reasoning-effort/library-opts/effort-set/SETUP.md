# Scenario

**Feature**: non-empty ModelReasoningEffort reaches RunSession opts

```
Opts.ModelReasoningEffort="max" -> AutoSendOrResume -> capture "max"
```

## Steps

1. Set Effort=`max`, Model empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-lib-effort"
	req.Model = ""
	req.ModelReasoningEffort = fixtureEffortMax // "max"
	return nil
}
```
