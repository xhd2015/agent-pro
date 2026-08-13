# Scenario

**Feature**: empty prompt is an API error

```
Run(Prompt="") -> error mentioning prompt
```

## Steps

1. Clear prompt after root default.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = ""
	req.InstallLaunch = true
	req.InstallWait = true
	return nil
}
```
