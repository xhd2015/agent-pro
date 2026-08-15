# Scenario

**Feature**: grok session parent help names focus

```
User -> parent session help line -> includes focus
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ParentHelp = true
	return nil
}
```
