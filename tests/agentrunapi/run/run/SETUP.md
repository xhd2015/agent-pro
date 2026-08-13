# Scenario

**Feature**: `Run` launch + wait + lifetime flags

```
Run(...) -> detach or open-terminal; SoftExit per KeepAliveDetached / ExitOnFinishTerminal
```

## Steps

1. Set mode `run`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "run"
	return nil
}
```
