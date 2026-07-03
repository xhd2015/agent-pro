# Scenario

**Feature**: parse button returned line from display alert

```
button returned:Yes — window opened -> button label extracted
```

## Steps

1. Feed stdout containing a single `button returned:` line.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Input = "button returned:Yes — window opened\n"
	return nil
}
```
