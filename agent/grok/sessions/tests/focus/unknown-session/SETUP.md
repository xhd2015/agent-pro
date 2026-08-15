# Scenario

**Feature**: unknown Grok session id is not found

```
empty sessions tree
-> focus unknown-id
-> not found
# Discover is not called
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee01"
	req.Args = []string{req.SessionID}
	return nil
}
```
