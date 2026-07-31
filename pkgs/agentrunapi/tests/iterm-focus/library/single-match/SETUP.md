# Scenario

**Feature**: unique iTerm candidate from serve ?? + ancestor real TTY

```
registry pid=serve(??) + ancestor /dev/ttys148
  -> FindByTTY one ref -> FocusSession policy single
```

## Steps

1. Apply `singleMatchFixtures` (serve ??, ancestor ttys148, one matching iTerm).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	singleMatchFixtures(req)
	return nil
}
```
