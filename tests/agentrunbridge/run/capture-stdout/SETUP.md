# Scenario

**Feature**: CaptureStdout returns trimmed launch stdout

```
Run(CaptureStdout=true, LaunchStdout="  hello out\\n") -> Result.Stdout == "hello out"
```

## Preconditions

- CaptureStdout true; WaitReady false.
- Scripted launch stdout has leading/trailing whitespace.

## Steps

1. Set CaptureStdout and LaunchStdout fixture.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "capture me"
	req.SessionID = "sess-capture"
	req.KeepTTY = true
	req.CaptureStdout = true
	req.WaitReady = false
	req.LaunchStdout = "  hello out\n"
	return nil
}
```
