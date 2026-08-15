# Scenario

**Feature**: missing session id is a usage error

```
User -> focus (no id) -> usage error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = nil
	return nil
}
```
