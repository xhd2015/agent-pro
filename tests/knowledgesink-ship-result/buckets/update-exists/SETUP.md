# Scenario

**Feature**: update bucket accepts an existing file on disk

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msg, branch := baseMsgBranch()
	req.SeedFiles = map[string]string{"SINK.md": "# sink\n"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		GitCommitMsg:  msg,
		GitBranchName: branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{
			Update: []string{"SINK.md"},
		},
	})
	if err != nil {
		return err
	}
	req.ResultJSON = body
	req.ExpectOK = true
	return nil
}
```
