# Scenario

**Feature**: corrupt config.json makes --show-config fail hard

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeConfigJSON(t, req.ConfigHome, "{not-json\n")
	req.Args = []string{"--show-config"}
	return nil
}
```
