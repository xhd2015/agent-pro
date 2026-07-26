# Scenario

**Feature**: append then list preserves message order and text

```
AppendMessage("a") + AppendMessage("b") -> ListMessages -> [a, b]
```

## Preconditions

- Session message queue starts empty.
- Two distinct message texts are appended.

## Steps

1. Set `req.Action = "append_list"`.
2. Set `req.MessageText` to `["first msg", "second msg"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "append_list"
	req.MessageText = []string{"first msg", "second msg"}
	return nil
}
```