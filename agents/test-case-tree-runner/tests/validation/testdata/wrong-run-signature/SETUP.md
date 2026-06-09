```go
import "testing"

type Request struct {
	Value int
}

type Response struct {
	Result int
}

func Run(t *testing.T, req *Request) error {
	return nil
}
```
