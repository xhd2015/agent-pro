# Scenario

**Feature**: second instance rejected while lock held

```
first listen acquires lock -> second listen -> exit non-zero with singleton message
```

## Steps

1. Inherit lock grouping setup (daemon + second instance).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
