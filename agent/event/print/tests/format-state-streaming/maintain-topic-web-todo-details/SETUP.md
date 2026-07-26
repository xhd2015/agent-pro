# Scenario

**Feature**: web/todo tool_call events show input details below the header

## Preconditions
- **Reproduces**: maintain-topic `--print` session lines for web fetch, Confluence
  search, and todo plan updates all render as bare tool headers.
- Uses exact `tool_call` events from session
  `20260625-114542-...-credit.pricing.center`.

## Steps
1. Set `req.Lines` to skynet search, todowrite, and webfetch events from the session.
2. Feed each line through `FormatState.FormatLine` like `FollowEventLog` does.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Lines = []string{
		`{"id":"prt_efce2959a001xX1PaK6UfeNeM9","type":"tool_call","timestamp":1782359168753,"tool":"skynet-base_get_doc_content","tool_input":{"doc":"https://fake.xhd2015.xyz/search?text=credit+pricing+center"},"mock":{"output":"Error: Authentication required. Please run 'skynet auth login' command and try again.","exit_code":0}}`,
		`{"id":"prt_efce28a9b001ZZQYXMiYq0mQpS","type":"tool_call","timestamp":1782359166423,"tool":"todowrite","tool_input":{"todos":[{"content":"搜索 Confluence 上 credit.pricing.center 相关文档","priority":"high","status":"in_progress"},{"content":"搜索 git 仓库了解项目信息","priority":"high","status":"pending"}]},"mock":{"exit_code":0}}`,
		`{"id":"prt_efce2f366001Xj2lBC5860JCtY","type":"tool_call","timestamp":1782359193157,"tool":"webfetch","tool_input":{"url":"https://fake.xhd2015.xyz/pages/viewpage.action?pageId=830343951"},"mock":{"exit_code":0}}`,
	}
	return nil
}
```