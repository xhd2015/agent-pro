# Scenario

**Feature**: non-empty Model reaches RunSession opts

```
Opts.Model="o3" -> AutoSendOrResume -> capture "o3"
```

## Steps

1. Set Model=`o3`, Effort empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-lib-model"
	req.Model = fixtureModel // "o3"
	req.ModelReasoningEffort = ""
	return nil
}
```
