# Scenario

**Feature**: RunJSON errors when the result file is not JSON

```
Wait writes "not-json" -> error
```

## Steps

1. WaitWriteRaw is not valid JSON.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.WaitWriteRaw = "not-json"
	return nil
}
```
