# Scenario

**Feature**: --color grays the notice: prefix when runner comes from config

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeConfigJSON(t, req.ConfigHome, "{\n  \"version\": 1,\n  \"agent_runner\": \"codex\"\n}\n")
	req.Args = []string{"--color", "hello color notice"}
	return nil
}
```
