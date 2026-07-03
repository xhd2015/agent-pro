# Scenario

**Feature**: parse text returned line from display dialog (Customize step)

```
text returned:VS Code opened but wrong workspace -> free-text answer extracted
```

## Steps

1. Feed stdout containing a `text returned:` line (second step of Customize flow).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Input = "text returned:VS Code opened but wrong workspace\n"
	return nil
}
```
