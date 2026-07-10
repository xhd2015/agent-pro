# Scenario

**Feature**: `--topic add-missing-scope --help` same as help-then-topic

```
Caller -> slack-msg --topic add-missing-scope --help -> same topic body -> exit 0
```

## Steps

1. Args `["--topic", "add-missing-scope", "--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--topic", "add-missing-scope", "--help"}
	return nil
}
```
