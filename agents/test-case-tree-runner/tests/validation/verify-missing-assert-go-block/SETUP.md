## Steps
- Point InputDir to the missing-assert-go-block fixture (ASSERT.md has no Go block)

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "missing-assert-go-block")
	return nil
}
```
