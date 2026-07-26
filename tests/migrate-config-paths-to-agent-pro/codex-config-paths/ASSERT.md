## Expected

- `DefaultConfigPath()` is `$HOME/.codex/config.toml`

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(req.Home, ".codex", "config.toml")
	if len(resp.Paths) != 1 || resp.Paths[0] != want {
		t.Fatalf("DefaultConfigPath() = %v, want [%s]", resp.Paths, want)
	}
}
```
