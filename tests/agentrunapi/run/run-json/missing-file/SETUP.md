# Scenario

**Feature**: RunJSON errors when the result file is missing / empty

```
Wait writes nothing; file left empty -> error mentions result path
```

## Steps

1. Skip Wait write so the reserved empty temp file stays invalid JSON.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.SkipWaitWrite = true
	return nil
}
```
