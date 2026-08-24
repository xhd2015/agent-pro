# Scenario

```go
import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	out := filepath.Join(req.TempDir, "pane.txt")
	req.Args = []string{req.SessionID, "-o", out}
	return nil
}
```
