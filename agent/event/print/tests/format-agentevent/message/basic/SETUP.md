## Preconditions
- A message event with text content.

## Steps
1. Set `req.Text = "hello world"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Text = "hello world"
	return nil
}
```
