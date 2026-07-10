# Scenario

**Feature**: unknown help topic name is rejected

```
Caller -> slack-msg --help --topic not-a-topic -> stderr unknown help topic -> exit 1
```

## Steps

1. Args `["--help", "--topic", "not-a-topic"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help", "--topic", "not-a-topic"}
	return nil
}
```
