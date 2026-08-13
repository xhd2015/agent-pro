# Scenario

**Feature**: RunJSON returns the result file JSON string

```
Wait writes {"a":1} -> RunJSON returns that string
```

## Steps

1. Launch + Wait write a known JSON object.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.WaitWriteJSON = `{"a":1}`
	return nil
}
```
