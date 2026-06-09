## Steps
- Point InputDir to the multiple-violations fixture (root missing Setup, leaf1 missing Setup + Assert, leaf2 no Run)

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "multiple-violations")
	return nil
}
```
