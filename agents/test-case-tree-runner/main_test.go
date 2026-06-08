package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTreeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func setupDoc(code string) string {
	return "# Setup\n\nAny section names are allowed.\n\n```go\n" + strings.TrimSpace(code) + "\n```\n"
}

func assertDoc(code string) string {
	return "# Assert\n\nAny section names are allowed.\n\n```go\n" + strings.TrimSpace(code) + "\n```\n"
}

func TestExtractFinalGoBlockIgnoresSectionNames(t *testing.T) {
	block, err := ExtractFinalGoBlock("SETUP.md", "## Anything\n\ntext\n\n```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n")
	if err != nil {
		t.Fatalf("extract final go block: %v", err)
	}
	if !strings.Contains(block.Code, "func Setup") {
		t.Fatalf("expected setup code, got %q", block.Code)
	}
}

func TestSetupRejectsMultipleGoBlocks(t *testing.T) {
	_, err := ParseSetupDocument("SETUP.md", "```go\nvar x = 1\n```\n\n```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n")
	if err == nil {
		t.Fatal("expected error for multiple go blocks")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("expected multiple go blocks error, got %v", err)
	}
}

func TestSetupRejectsNonFinalGoBlock(t *testing.T) {
	_, err := ParseSetupDocument("SETUP.md", "```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n\nmore markdown\n")
	if err == nil {
		t.Fatal("expected error for non-final go block")
	}
	if !strings.Contains(err.Error(), "final") {
		t.Fatalf("expected final go block error, got %v", err)
	}
}

func TestAssertRequiresSingleFinalGoBlock(t *testing.T) {
	_, err := ParseAssertDocument("ASSERT.md", "## Expected\n- no executable assertion\n")
	if err == nil {
		t.Fatal("expected error for missing assert go block")
	}
	if !strings.Contains(err.Error(), "go block") {
		t.Fatalf("expected go block error, got %v", err)
	}
}

func TestRootSetupRequiresRequestAndResponseTypes(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
func Run(t *testing.T, req *Request) (*Response, error) { return nil, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	_, err := DiscoverTreeCases(root)
	if err == nil {
		t.Fatal("expected missing Request/Response error")
	}
	if !strings.Contains(err.Error(), "Request") || !strings.Contains(err.Error(), "Response") {
		t.Fatalf("expected Request/Response error, got %v", err)
	}
}

func TestSetupFunctionSignature(t *testing.T) {
	_, err := ParseSetupDocument("SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Setup(t *testing.T) {}
`))
	if err == nil {
		t.Fatal("expected invalid setup signature error")
	}
	if !strings.Contains(err.Error(), "Setup") || !strings.Contains(err.Error(), "*Request") {
		t.Fatalf("expected setup signature error, got %v", err)
	}
}

func TestRunFunctionSignature(t *testing.T) {
	_, err := ParseSetupDocument("SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) error { return nil }
`))
	if err == nil {
		t.Fatal("expected invalid run signature error")
	}
	if !strings.Contains(err.Error(), "Run") || !strings.Contains(err.Error(), "*Response") {
		t.Fatalf("expected run signature error, got %v", err)
	}
}

func TestAssertFunctionSignature(t *testing.T) {
	_, err := ParseAssertDocument("ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, err error) {}
`))
	if err == nil {
		t.Fatal("expected invalid assert signature error")
	}
	if !strings.Contains(err.Error(), "Assert") || !strings.Contains(err.Error(), "*Response") {
		t.Fatalf("expected assert signature error, got %v", err)
	}
}

func TestAtLeastOneRunRequiredInSetupChain(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	err := RunTree(root)
	if err == nil {
		t.Fatal("expected missing Run error")
	}
	if !strings.Contains(err.Error(), "Run") {
		t.Fatalf("expected missing Run error, got %v", err)
	}
}

func TestDeepestRunOverridesAncestorRun(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{ Source string }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Source: "root"}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Source: "leaf"}, nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp.Source != "leaf" { t.Fatalf("expected leaf run, got %q", resp.Source) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestExecutionOrderSetupRunAssert(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{ Order []string }
type Response struct{}
func Setup(t *testing.T, req *Request) error { req.Order = append(req.Order, "root setup"); return nil }
func Run(t *testing.T, req *Request) (*Response, error) { req.Order = append(req.Order, "run"); return &Response{}, nil }
`))
	writeTreeFile(t, root, "parent/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Order = append(req.Order, "parent setup"); return nil }
`))
	writeTreeFile(t, root, "parent/leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Order = append(req.Order, "leaf setup"); return nil }
`))
	writeTreeFile(t, root, "parent/leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	req.Order = append(req.Order, "assert")
	want := []string{"root setup", "parent setup", "leaf setup", "run", "assert"}
	if !reflect.DeepEqual(req.Order, want) { t.Fatalf("order = %#v, want %#v", req.Order, want) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestSetupErrorFailsBeforeRunAndAssert(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Setup(t *testing.T, req *Request) error { return fmt.Errorf("setup failed") }
func Run(t *testing.T, req *Request) (*Response, error) { t.Fatal("run should not execute"); return nil, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Fatal("assert should not execute")
}
`))

	err := RunTree(root)
	if err == nil {
		t.Fatal("expected setup failure")
	}
	if !strings.Contains(err.Error(), "setup failed") {
		t.Fatalf("expected setup failure in error, got %v", err)
	}
}

func TestRunErrorPassedToAssert(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return nil, fmt.Errorf("run failed") }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil || err.Error() != "run failed" { t.Fatalf("expected run failed error, got %v", err) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestRequestMutatedThroughSetupChain(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{ Value int }
type Response struct{ Value int }
func Setup(t *testing.T, req *Request) error { req.Value += 1; return nil }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Value: req.Value}, nil }
`))
	writeTreeFile(t, root, "parent/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Value += 2; return nil }
`))
	writeTreeFile(t, root, "parent/leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Value += 3; return nil }
`))
	writeTreeFile(t, root, "parent/leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if req.Value != 6 || resp.Value != 6 { t.Fatalf("req=%d resp=%d, want 6", req.Value, resp.Value) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestResponsePassedFromRunToAssert(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{ Message string }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Message: "ok"}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp == nil || resp.Message != "ok" { t.Fatalf("unexpected response: %#v", resp) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestDuplicateSetupHooksAcrossLevelsAllowed(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{ Count int }
type Response struct{}
func Setup(t *testing.T, req *Request) error { req.Count++; return nil }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "parent/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Count++; return nil }
`))
	writeTreeFile(t, root, "parent/leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Count++; return nil }
`))
	writeTreeFile(t, root, "parent/leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if req.Count != 3 { t.Fatalf("count = %d, want 3", req.Count) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestDuplicateNamesAcrossDifferentLeavesAllowed(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{ Name string }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Name: "root"}, nil }
`))
	writeTreeFile(t, root, "a/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "a/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	var duplicate = "a"
	if duplicate != "a" { t.Fatal(duplicate) }
}
`))
	writeTreeFile(t, root, "b/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "b/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	var duplicate = "b"
	if duplicate != "b" { t.Fatal(duplicate) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestChildCannotRedefineRequestOrResponse(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
type Request struct{ Bad bool }
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	err := RunTree(root)
	if err == nil {
		t.Fatal("expected child Request redefinition error")
	}
	if !strings.Contains(err.Error(), "Request") {
		t.Fatalf("expected Request conflict error, got %v", err)
	}
}

func TestFailingAssertReturnsLeafPath(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Fatal("assert failed")
}
`))

	err := RunTree(root)
	if err == nil {
		t.Fatal("expected assert failure")
	}
	if !strings.Contains(err.Error(), "leaf") || !strings.Contains(err.Error(), "assert failed") {
		t.Fatalf("expected leaf path and assert failure, got %v", err)
	}
}

func TestInvalidGoReportsSourceMarkdown(t *testing.T) {
	_, err := ParseSetupDocument("branch/SETUP.md", setupDoc(`
type Request struct{
`))
	if err == nil {
		t.Fatal("expected invalid go error")
	}
	if !strings.Contains(err.Error(), "branch/SETUP.md") {
		t.Fatalf("expected source markdown path in error, got %v", err)
	}
}

func TestImportsResolvedByGoimports(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{ Name string }
type Response struct{ Message string }
func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Message: fmt.Sprintf("hello %s", req.Name)}, nil
}
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { req.Name = "world"; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp.Message != "hello world" { t.Fatal(resp.Message) }
}
`))

	if err := RunTree(root); err != nil {
		t.Fatalf("run tree: %v", err)
	}
}

func TestGenerateCodeScaffoldsProseOnlyTreeAndCompiles(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", "# Setup\n\nRoot model goes here.\n")
	writeTreeFile(t, root, "leaf/SETUP.md", "# Setup\n\nLeaf-specific setup.\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n\nLeaf assertion.\n")

	if err := GenerateCode(root, GenerateCodeOptions{}); err != nil {
		t.Fatalf("generate code: %v", err)
	}

	rootSetup, err := os.ReadFile(filepath.Join(root, "SETUP.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rootSetup), "type Request struct") ||
		!strings.Contains(string(rootSetup), "type Response struct") ||
		!strings.Contains(string(rootSetup), "func Run(t *testing.T, req *Request) (*Response, error)") {
		t.Fatalf("root setup missing scaffold:\n%s", rootSetup)
	}

	leafSetup, err := os.ReadFile(filepath.Join(root, "leaf/SETUP.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(leafSetup), "func Setup(t *testing.T, req *Request) error") {
		t.Fatalf("leaf setup missing scaffold:\n%s", leafSetup)
	}

	leafAssert, err := os.ReadFile(filepath.Join(root, "leaf/ASSERT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(leafAssert), "func Assert(t *testing.T, req *Request, resp *Response, err error)") {
		t.Fatalf("leaf assert missing scaffold:\n%s", leafAssert)
	}

	if err := CompileTree(root); err != nil {
		t.Fatalf("compile generated tree: %v", err)
	}
}

func TestGenerateCodeCreatesMissingLeafSetup(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", "# Setup\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n")

	if err := GenerateCode(root, GenerateCodeOptions{}); err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "leaf/SETUP.md")); err != nil {
		t.Fatalf("expected generated leaf SETUP.md: %v", err)
	}
}

func TestGenerateCodePreservesExistingGoBlocks(t *testing.T) {
	root := t.TempDir()
	rootCode := `
type Request struct{ Name string }
type Response struct{ Message string }
func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Message: "ok"}, nil
}
`
	assertCode := `
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
}
`
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(rootCode))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`func Setup(t *testing.T, req *Request) error { req.Name = "kept"; return nil }`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(assertCode))

	beforeRoot, _ := os.ReadFile(filepath.Join(root, "SETUP.md"))
	beforeAssert, _ := os.ReadFile(filepath.Join(root, "leaf/ASSERT.md"))
	if err := GenerateCode(root, GenerateCodeOptions{}); err != nil {
		t.Fatalf("generate code: %v", err)
	}
	afterRoot, _ := os.ReadFile(filepath.Join(root, "SETUP.md"))
	afterAssert, _ := os.ReadFile(filepath.Join(root, "leaf/ASSERT.md"))
	if string(afterRoot) != string(beforeRoot) {
		t.Fatalf("root setup changed unexpectedly:\n%s", afterRoot)
	}
	if string(afterAssert) != string(beforeAssert) {
		t.Fatalf("assert changed unexpectedly:\n%s", afterAssert)
	}
}

func TestGenerateCodeDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", "# Setup\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n")

	if err := GenerateCode(root, GenerateCodeOptions{DryRun: true}); err != nil {
		t.Fatalf("dry-run generate code: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "SETUP.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "type Request struct") {
		t.Fatalf("dry run wrote root setup:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(root, "leaf/SETUP.md")); !os.IsNotExist(err) {
		t.Fatalf("dry run should not create leaf setup, stat err=%v", err)
	}
}

func TestGenerateCodeRejectsInvalidExistingNonFinalGoBlock(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", "```go\nfunc Setup(t *testing.T, req *Request) error { return nil }\n```\n\nmore markdown\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n")

	err := GenerateCode(root, GenerateCodeOptions{})
	if err == nil {
		t.Fatal("expected non-final go block error")
	}
	if !strings.Contains(err.Error(), "final") {
		t.Fatalf("expected final block error, got %v", err)
	}
}

func TestGenerateCodeRuntimeRemainsRedForDefaultRun(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", "# Setup\n")
	writeTreeFile(t, root, "leaf/SETUP.md", "# Setup\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n")

	if err := GenerateCode(root, GenerateCodeOptions{}); err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := CompileTree(root); err != nil {
		t.Fatalf("compile generated tree: %v", err)
	}
	err := RunTree(root)
	if err == nil {
		t.Fatal("expected runtime RED failure")
	}
	if !strings.Contains(err.Error(), "not implemented yet") {
		t.Fatalf("expected not implemented error, got %v", err)
	}
}
