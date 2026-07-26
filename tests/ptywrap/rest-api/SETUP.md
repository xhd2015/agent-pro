# Scenario

**Feature**: REST session management API

```
# HTTP CRUD
POST /api/terminal/sessions -> create
PATCH /api/terminal/sessions/{id} -> rename
GET /api/terminal/sessions -> list
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "rest-create"
	return nil
}
```