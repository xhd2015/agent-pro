# Scenario

**Feature**: `RunJSON` schema example + temp result file

```
RunJSON(prompt, SchemaExample) -> append path+example; read JSON
```

## Steps

1. Set mode `run_json` and a default schema example.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "run_json"
	if req.SchemaExample == "" {
		req.SchemaExample = `{"ok":true,"url":"https://example.com"}`
	}
	return nil
}
```
