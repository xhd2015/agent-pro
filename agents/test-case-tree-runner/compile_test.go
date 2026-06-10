package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func TestRootSetupRequiresRequestAndResponseTypes(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
func Run(t *testing.T, req *Request) (*Response, error) { return nil, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	_, err := core.DiscoverTreeCases(root)
	if err == nil {
		t.Fatal("expected missing Request/Response error")
	}
	if !strings.Contains(err.Error(), "Request") || !strings.Contains(err.Error(), "Response") {
		t.Fatalf("expected Request/Response error, got %v", err)
	}
}

func TestAtLeastOneRunRequiredInSetupChain(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	err := TestTree(root, core.Options{})
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

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Fatal("assert should not execute")
}
`))

	err := TestTree(root, core.Options{})
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil || err.Error() != "run failed" { t.Fatalf("expected run failed error, got %v", err) }
}
`))

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp == nil || resp.Message != "ok" { t.Fatalf("unexpected response: %#v", resp) }
}
`))

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "a/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	var duplicate = "a"
	if duplicate != "a" { t.Fatal(duplicate) }
}
`))
	writeTreeFile(t, root, "b/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "b/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	var duplicate = "b"
	if duplicate != "b" { t.Fatal(duplicate) }
}
`))

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	err := TestTree(root, core.Options{})
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
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Fatal("assert failed")
}
`))

	err := TestTree(root, core.Options{})
	if err == nil {
		t.Fatal("expected assert failure")
	}
	if !strings.Contains(err.Error(), "leaf") || !strings.Contains(err.Error(), "assert failed") {
		t.Fatalf("expected leaf path and assert failure, got %v", err)
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

	if err := TestTree(root, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
}

func TestBuildWithGenDirWritesGeneratedCode(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build with gen dir: %v", err)
	}
	generated := filepath.Join(genDir, "leaf_test.go")
	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("read generated test: %v", err)
	}
	if !strings.Contains(string(data), "func TestGeneratedCaseLeaf") {
		t.Fatalf("generated test missing TestGeneratedCaseLeaf:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(genDir, "leaf")); !os.IsNotExist(err) {
		t.Fatalf("expected no per-test directory, stat err=%v", err)
	}
}

func TestCompileNonVerboseOutputFormat(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir, Stderr: &stderr}); err != nil {
		t.Fatalf("build: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "doctest:") {
		t.Fatalf("expected source dir in output, got %q", out)
	}
	if !strings.Contains(out, "─── 1 test cases") {
		t.Fatalf("expected test case count, got %q", out)
	}
	if !strings.Contains(out, genDir) {
		t.Fatalf("expected gen dir %q in output, got %q", genDir, out)
	}
	if !strings.Contains(out, "cd ") || !strings.Contains(out, "go build") {
		t.Fatalf("expected go build command in output, got %q", out)
	}
}

func TestCompileVerboseStreamingOutput(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir, Verbose: true, Stderr: &stderr}); err != nil {
		t.Fatalf("build verbose: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "SETUP.md") {
		t.Fatalf("expected SETUP.md info, got %q", out)
	}
	if !strings.Contains(out, "Request, Response, Setup, Run") {
		t.Fatalf("expected Request/Response/Run types, got %q", out)
	}
	if !strings.Contains(out, "defines Run") {
		t.Fatalf("expected 'defines Run' annotation, got %q", out)
	}
	if !strings.Contains(out, "✦") {
		t.Fatalf("expected leaf marker (✦), got %q", out)
	}
	if !strings.Contains(out, "(Run:") {
		t.Fatalf("expected Run source annotation, got %q", out)
	}
	if !strings.Contains(out, "leaf/SETUP.md") {
		t.Fatalf("expected leaf SETUP.md info, got %q", out)
	}
	if !strings.Contains(out, "─── 1 test cases") {
		t.Fatalf("expected test case count, got %q", out)
	}
	if !strings.Contains(out, genDir) {
		t.Fatalf("expected gen dir %q in output, got %q", genDir, out)
	}
	if !strings.Contains(out, "cd ") || !strings.Contains(out, "go build") || !strings.Contains(out, "-v") {
		t.Fatalf("expected go build -v command in output, got %q", out)
	}
}

func TestCompileVerboseShowsRunOverride(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "child/SETUP.md", setupDoc(`
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "child/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir, Verbose: true, Stderr: &stderr}); err != nil {
		t.Fatalf("build verbose: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "child/SETUP.md") {
		t.Fatalf("expected child SETUP.md, got %q", out)
	}
	if !strings.Contains(out, "(Run: child/SETUP.md)") {
		t.Fatalf("expected Run source from child, got %q", out)
	}
}

func TestCompileGoModGeneratedWithModule(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/test\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "tests/README.md", "# tree")
	writeTreeFile(t, srcRoot, "tests/SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, srcRoot, "tests/leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, srcRoot, "tests/leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(filepath.Join(srcRoot, "tests"), core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	goModData, err := os.ReadFile(filepath.Join(genDir, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod in gen dir: %v", err)
	}
	goMod := string(goModData)
	if !strings.Contains(goMod, "module testcase") {
		t.Fatalf("expected module testcase, got %q", goMod)
	}
	if !strings.Contains(goMod, "example.com/test") {
		t.Fatalf("expected require example.com/test, got %q", goMod)
	}
	if !strings.Contains(goMod, "replace example.com/test => "+srcRoot) {
		t.Fatalf("expected replace with abs path %q, got %q", srcRoot, goMod)
	}
}

func TestCompileNoGoModWhenNoSourceModule(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	goModData, err := os.ReadFile(filepath.Join(genDir, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod in gen dir: %v", err)
	}
	goMod := string(goModData)
	if !strings.Contains(goMod, "module testcase") {
		t.Fatalf("expected module testcase, got %q", goMod)
	}
	if strings.Contains(goMod, "replace ") {
		t.Fatalf("expected no replace directive when no source module, got %q", goMod)
	}
}

func TestCompileDefaultKeepsTempDir(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := build.Build(root, core.Options{Stderr: &stderr}); err != nil {
		t.Fatalf("build: %v", err)
	}

	out := stderr.String()
	i := strings.Index(out, "→ ")
	if i < 0 {
		t.Fatalf("expected → gen dir in output, got %q", out)
	}
	rest := out[i+len("→ "):]
	genDir := strings.TrimSpace(strings.SplitN(rest, "\n", 2)[0])

	if _, err := os.Stat(genDir); os.IsNotExist(err) {
		t.Fatalf("expected temp dir %s to be kept, but it was removed", genDir)
	}
	defer os.RemoveAll(genDir)
}

func TestCompileRmRemovesTempDir(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := build.Build(root, core.Options{RemoveTemp: true, Stderr: &stderr}); err != nil {
		t.Fatalf("build: %v", err)
	}

	out := stderr.String()
	i := strings.Index(out, "→ ")
	if i < 0 {
		t.Fatalf("expected → gen dir in output, got %q", out)
	}
	rest := out[i+len("→ "):]
	genDir := strings.TrimSpace(strings.SplitN(rest, "\n", 2)[0])

	if _, err := os.Stat(genDir); !os.IsNotExist(err) {
		t.Fatalf("expected temp dir %s to be removed, but it still exists", genDir)
	}
}

func TestCompileRmDoesNotRemoveGenDir(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir, RemoveTemp: true}); err != nil {
		t.Fatalf("build: %v", err)
	}

	if _, err := os.Stat(genDir); os.IsNotExist(err) {
		t.Fatalf("expected gen dir %s to exist with --rm, but it was removed", genDir)
	}
}

func TestCompileWithExternalPackageImport(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/mylib\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "pkg/helper.go", "package pkg\n\nfunc Name() string { return \"hello\" }\n")

	testsDir := filepath.Join(srcRoot, "tests")
	writeTreeFile(t, testsDir, "README.md", "# tree")
	writeTreeFile(t, testsDir, "SETUP.md", setupDoc(`
import "example.com/mylib/pkg"
type Request struct{}
type Response struct{ Val string }
func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Val: pkg.Name()}, nil
}
`))
	writeTreeFile(t, testsDir, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, testsDir, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp.Val != "hello" { t.Fatalf("expected hello, got %q", resp.Val) }
}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(testsDir, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	goModData, err := os.ReadFile(filepath.Join(genDir, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod in gen dir: %v", err)
	}
	goMod := string(goModData)
	if !strings.Contains(goMod, "replace example.com/mylib => "+srcRoot) {
		t.Fatalf("expected replace with path %q, got %q", srcRoot, goMod)
	}

	cmd := exec.Command("go", "test", "-count=1", "./...")
	cmd.Dir = genDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go test in gen dir failed: %v\n%s", err, out)
	}
}

func TestDiscoverTreeCasesVerbose(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "setup_leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "setup_leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var buf bytes.Buffer
	cases, err := core.DiscoverTreeCasesVerbose(root, &buf)
	if err != nil {
		t.Fatalf("discover verbose: %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(cases))
	}
	out := buf.String()
	if !strings.Contains(out, "SETUP.md") {
		t.Fatalf("expected root SETUP, got %q", out)
	}
	if !strings.Contains(out, "setup_leaf/SETUP.md") {
		t.Fatalf("expected leaf SETUP, got %q", out)
	}
	if !strings.Contains(out, "✦") {
		t.Fatalf("expected leaf marker, got %q", out)
	}
}

func TestFindModuleRoot(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/a\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "a/b/c/SETUP.md", "# test\n")

	modRoot, modPath, ok := core.FindModuleRoot(filepath.Join(srcRoot, "a/b/c"))
	if !ok {
		t.Fatal("expected to find module root")
	}
	if modRoot != srcRoot {
		t.Fatalf("expected mod root %q, got %q", srcRoot, modRoot)
	}
	if modPath != "example.com/a" {
		t.Fatalf("expected mod path example.com/a, got %q", modPath)
	}
}

func TestFindModuleRootNotFound(t *testing.T) {
	root := t.TempDir()
	_, _, ok := core.FindModuleRoot(root)
	if ok {
		t.Fatal("expected no module root")
	}
}

func TestSourceFilesCopiedWithPackageRename(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/mylib\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "pkg/helper.go", "package pkg\n\nfunc Exported() string { return \"exported\" }\nfunc private() string { return \"private\" }\n")
	writeTreeFile(t, srcRoot, "pkg/types.go", "package pkg\n\ntype MyType struct{ Val string }\n")

	testsDir := filepath.Join(srcRoot, "tests")
	writeTreeFile(t, testsDir, "README.md", "# tree")
	writeTreeFile(t, testsDir, "SETUP.md", strings.ReplaceAll(`# Setup
- Go module: example.com/mylib
- Package under test: pkg

`+"```go\n"+
		`type Request struct{}
type Response struct{ Msg string }
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Msg: Exported() + " " + private()}, nil
}
`+"```\n", "\n", "\n"))

	writeTreeFile(t, testsDir, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, testsDir, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp.Msg != "exported private" { t.Fatalf("got %q", resp.Msg) }
}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(testsDir, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	helperPath := filepath.Join(genDir, "helper.go")
	helperData, err := os.ReadFile(helperPath)
	if err != nil {
		t.Fatalf("expected helper.go in gen dir: %v", err)
	}
	if !strings.Contains(string(helperData), "package pkg_tc") {
		t.Fatalf("expected package pkg_tc in helper.go, got:\n%s", helperData)
	}
	if strings.Contains(string(helperData), "package pkg\n") {
		t.Fatal("expected original package name to be replaced")
	}
	if !strings.Contains(string(helperData), "func Exported()") {
		t.Fatal("expected Exported func in copied helper.go")
	}
	if !strings.Contains(string(helperData), "func private()") {
		t.Fatal("expected private func in copied helper.go")
	}

	typesPath := filepath.Join(genDir, "types.go")
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Fatal("expected types.go in gen dir")
	}

	leafTestPath := filepath.Join(genDir, "leaf_test.go")
	testData, err := os.ReadFile(leafTestPath)
	if err != nil {
		t.Fatalf("expected leaf_test.go: %v", err)
	}
	if !strings.Contains(string(testData), "package pkg_tc") {
		t.Fatalf("expected package pkg_tc in test file, got:\n%s", testData)
	}

	goModData, err := os.ReadFile(filepath.Join(genDir, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod: %v", err)
	}
	if !strings.Contains(string(goModData), "module testcase") {
		t.Fatalf("expected module testcase in go.mod, got %s", goModData)
	}
}

func TestPrivateFunctionAccessibleInTest(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/mylib\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "pkg/helper.go", "package pkg\n\nfunc getSecret() string { return \"top-secret\" }\n")

	testsDir := filepath.Join(srcRoot, "tests")
	writeTreeFile(t, testsDir, "README.md", "# tree")
	writeTreeFile(t, testsDir, "SETUP.md", strings.ReplaceAll(`# Setup
- Go module: example.com/mylib
- Package under test: pkg

`+"```go\n"+
		`type Request struct{}
type Response struct{ Result string }
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Result: getSecret()}, nil
}
`+"```\n", "\n", "\n"))

	writeTreeFile(t, testsDir, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, testsDir, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil { t.Fatal(err) }
	if resp.Result != "top-secret" { t.Fatalf("got %q", resp.Result) }
}
`))

	if err := TestTree(testsDir, core.Options{}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
}

func TestNoTestGoFilesCopied(t *testing.T) {
	srcRoot := t.TempDir()
	writeTreeFile(t, srcRoot, "go.mod", "module example.com/mylib\n\ngo 1.21\n")
	writeTreeFile(t, srcRoot, "pkg/helper.go", "package pkg\n\nfunc Ok() string { return \"ok\" }\n")
	writeTreeFile(t, srcRoot, "pkg/helper_test.go", "package pkg\n\nimport \"testing\"\n\nfunc TestFoo(t *testing.T) {}\n")

	testsDir := filepath.Join(srcRoot, "tests")
	writeTreeFile(t, testsDir, "README.md", "# tree")
	writeTreeFile(t, testsDir, "SETUP.md", strings.ReplaceAll(`# Setup
- Go module: example.com/mylib
- Package under test: pkg

`+"```go\n"+
		`type Request struct{}
type Response struct{}
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`+"```\n", "\n", "\n"))

	writeTreeFile(t, testsDir, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, testsDir, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(testsDir, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(genDir, "helper.go")); os.IsNotExist(err) {
		t.Fatal("expected helper.go to be copied")
	}
	if _, err := os.Stat(filepath.Join(genDir, "helper_test.go")); !os.IsNotExist(err) {
		t.Fatal("expected helper_test.go NOT to be copied")
	}
}

func TestBackwardCompatNoPackageUnderTest(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	leafTestData, err := os.ReadFile(filepath.Join(genDir, "leaf_test.go"))
	if err != nil {
		t.Fatalf("read leaf_test.go: %v", err)
	}
	if !strings.Contains(string(leafTestData), "package testcase") {
		t.Fatalf("expected package testcase when no metadata, got:\n%s", leafTestData)
	}
}

func TestValidationReportsAllErrorsAtOnce(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "SETUP.md", "# Setup\n\n```go\nfunc Setup(t *testing.T, req *Request) error { _ = req; return nil }\n```\n")
	writeTreeFile(t, root, "leaf/SETUP.md", "# Setup\n\nProse only, no Go block.\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n\n```go\nfunc Check(t *testing.T, req *Request, resp *Response, err error) {}\n```\n")

	_, err := core.DiscoverTreeCases(root)
	if err == nil {
		t.Fatal("expected validation error")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "must define type Request") {
		t.Fatalf("expected missing types error, got %q", errStr)
	}
	if !strings.Contains(errStr, "must have a Go code block") {
		t.Fatalf("expected missing Go block error, got %q", errStr)
	}
	if !strings.Contains(errStr, "missing func Assert") {
		t.Fatalf("expected missing Assert error, got %q", errStr)
	}
	if !strings.Contains(errStr, "validation errors:") {
		t.Fatalf("expected validation errors header, got %q", errStr)
	}
	lines := strings.Split(strings.TrimSpace(errStr), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 3 errors, got %q", errStr)
	}
}

func TestValidationNoErrorForValidTree(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	if _, err := core.DiscoverTreeCases(root); err != nil {
		t.Fatalf("expected no error for valid tree, got %v", err)
	}
}

func TestValidationTestdataDirSkipped(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))
	writeTreeFile(t, root, "testdata/SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
`))
	writeTreeFile(t, root, "testdata/leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "testdata/leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	cases, err := core.DiscoverTreeCases(root)
	if err != nil {
		t.Fatalf("expected no error, testdata/ should be skipped: %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("expected 1 case (testdata/ skipped), got %d", len(cases))
	}
}

func TestRunVerboseFlag(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := TestTree(root, core.Options{Verbose: true, Stderr: &stderr}); err != nil {
		t.Fatalf("test tree verbose: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "TestGeneratedCaseLeaf") {
		t.Fatalf("expected test function name in output, got %q", out)
	}
}

func TestRunShowsGenDirAndCommand(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := TestTree(root, core.Options{Stderr: &stderr}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "→ ") {
		t.Fatalf("expected gen dir in output, got %q", out)
	}
	if !strings.Contains(out, "cd ") || !strings.Contains(out, "go test -c") {
		t.Fatalf("expected cd xxx && go test -c ... in output, got %q", out)
	}
}

func TestRunDefaultKeepsTempDir(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := TestTree(root, core.Options{Stderr: &stderr}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
	out := stderr.String()
	i := strings.Index(out, "→ ")
	if i < 0 {
		t.Fatalf("expected → gen dir in output, got %q", out)
	}
	rest := out[i+len("→ "):]
	genDir := strings.TrimSpace(strings.SplitN(rest, "\n", 2)[0])

	if _, err := os.Stat(genDir); os.IsNotExist(err) {
		t.Fatalf("expected temp dir %s to be kept, but it was removed", genDir)
	}
	defer os.RemoveAll(genDir)
}

func TestRunRmFlag(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := TestTree(root, core.Options{RemoveTemp: true, Stderr: &stderr}); err != nil {
		t.Fatalf("test tree: %v", err)
	}
	out := stderr.String()
	i := strings.Index(out, "→ ")
	if i < 0 {
		t.Fatalf("expected → gen dir in output, got %q", out)
	}
	rest := out[i+len("→ "):]
	genDir := strings.TrimSpace(strings.SplitN(rest, "\n", 2)[0])

	if _, err := os.Stat(genDir); !os.IsNotExist(err) {
		t.Fatalf("expected temp dir %s to be removed, but it still exists", genDir)
	}
}

func TestRunVerboseShowsPerCaseOutput(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "a/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "a/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))
	writeTreeFile(t, root, "b/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "b/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	if err := TestTree(root, core.Options{Verbose: true, Stderr: &stderr}); err != nil {
		t.Fatalf("test tree verbose: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "TestGeneratedCaseA") {
		t.Fatalf("expected TestGeneratedCaseA in output, got %q", out)
	}
	if !strings.Contains(out, "TestGeneratedCaseB") {
		t.Fatalf("expected TestGeneratedCaseB in output, got %q", out)
	}
}

func TestParseRunOptionsFlagAfterDir(t *testing.T) {
	opts, remainder, err := parseTestOptions([]string{"mydir", "-v", "--rm"})
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainder) != 1 || remainder[0] != "mydir" {
		t.Fatalf("expected remainder [mydir], got %v", remainder)
	}
	if !opts.Verbose {
		t.Fatal("expected Verbose=true")
	}
	if !opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=true")
	}
}

func TestParseRunOptionsFlagBeforeDir(t *testing.T) {
	opts, remainder, err := parseTestOptions([]string{"-v", "--rm", "mydir"})
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainder) != 1 || remainder[0] != "mydir" {
		t.Fatalf("expected remainder [mydir], got %v", remainder)
	}
	if !opts.Verbose {
		t.Fatal("expected Verbose=true")
	}
	if !opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=true")
	}
}

func TestParseBuildOptionsFlagAfterDir(t *testing.T) {
	opts, remainder, err := parseBuildOptions([]string{"mydir", "-v", "--rm", "--gen-dir", "/tmp/gen"})
	if err != nil {
		t.Fatalf("parseBuildOptions: %v", err)
	}
	if len(remainder) != 1 || remainder[0] != "mydir" {
		t.Fatalf("expected remainder [mydir], got %v", remainder)
	}
	if !opts.Verbose {
		t.Fatal("expected Verbose=true")
	}
	if !opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=true")
	}
	if opts.GenDir != "/tmp/gen" {
		t.Fatalf("expected GenDir=/tmp/gen, got %q", opts.GenDir)
	}
}

func TestParseBuildOptionsGenDirEquals(t *testing.T) {
	opts, remainder, err := parseBuildOptions([]string{"--gen-dir=/tmp/gen", "mydir"})
	if err != nil {
		t.Fatalf("parseBuildOptions: %v", err)
	}
	if opts.GenDir != "/tmp/gen" {
		t.Fatalf("expected GenDir=/tmp/gen, got %q", opts.GenDir)
	}
	if len(remainder) != 1 || remainder[0] != "mydir" {
		t.Fatalf("expected remainder [mydir], got %v", remainder)
	}
}

func TestParseBuildOptionsCountEquals(t *testing.T) {
	opts, remainder, err := parseBuildOptions([]string{"-count=5", "mydir"})
	if err != nil {
		t.Fatalf("parseBuildOptions: %v", err)
	}
	if opts.Count != 5 {
		t.Fatalf("expected Count=5, got %d", opts.Count)
	}
	if len(remainder) != 1 || remainder[0] != "mydir" {
		t.Fatalf("expected remainder [mydir], got %v", remainder)
	}
}

func TestRunFlagAfterDirIntegration(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	var stderr bytes.Buffer
	opts, remainArgs, err := parseTestOptions(append([]string{root}, "-v", "--rm"))
	if err != nil {
		t.Fatalf("parseTestOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != root {
		t.Fatalf("expected remainder [root], got %v", remainArgs)
	}
	if !opts.Verbose {
		t.Fatal("expected Verbose=true")
	}
	if !opts.RemoveTemp {
		t.Fatal("expected RemoveTemp=true")
	}
	opts.Stderr = &stderr
	if err := TestTree(remainArgs[0], opts); err != nil {
		t.Fatalf("test tree: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "TestGeneratedCaseLeaf") {
		t.Fatalf("expected TestGeneratedCaseLeaf in output, got %q", out)
	}
}

func TestCompileFlagAfterDirIntegration(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	var stderr bytes.Buffer
	opts, remainArgs, err := parseBuildOptions(append([]string{root}, "--gen-dir="+genDir, "-v"))
	if err != nil {
		t.Fatalf("parseBuildOptions: %v", err)
	}
	if len(remainArgs) != 1 || remainArgs[0] != root {
		t.Fatalf("expected remainder [root], got %v", remainArgs)
	}
	if !opts.Verbose {
		t.Fatal("expected Verbose=true")
	}
	if opts.GenDir != genDir {
		t.Fatalf("expected GenDir=%q, got %q", genDir, opts.GenDir)
	}
	opts.Stderr = &stderr
	if err := build.Build(remainArgs[0], opts); err != nil {
		t.Fatalf("build tree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(genDir, "leaf_test.go")); err != nil {
		t.Fatalf("expected generated leaf_test.go in gen dir: %v", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "─── 1 test cases") {
		t.Fatalf("expected test case count, got %q", out)
	}
}

func TestGeneratedCodeHasDoctestRootConst(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "leaf/SETUP.md", setupDoc(`
func Setup(t *testing.T, req *Request) error { _ = req; return nil }
`))
	writeTreeFile(t, root, "leaf/ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	leafTestData, err := os.ReadFile(filepath.Join(genDir, "leaf_test.go"))
	if err != nil {
		t.Fatalf("read leaf_test.go: %v", err)
	}
	code := string(leafTestData)

	absRoot, _ := filepath.Abs(root)
	if !strings.Contains(code, "const DOCTEST_ROOT = `"+absRoot+"`") {
		t.Fatalf("expected DOCTEST_ROOT const with path %q, got:\n%s", absRoot, code)
	}
	if !strings.Contains(code, "os.Chdir(filepath.Join(DOCTEST_ROOT, \"leaf\"))") {
		t.Fatalf("expected os.Chdir(filepath.Join(DOCTEST_ROOT, \"leaf\")), got:\n%s", code)
	}
	if !strings.Contains(code, "__origWd, __wdErr := os.Getwd()") {
		t.Fatalf("expected os.Getwd() before chdir, got:\n%s", code)
	}
	if !strings.Contains(code, "defer os.Chdir(__origWd)") {
		t.Fatalf("expected defer os.Chdir(__origWd), got:\n%s", code)
	}
	if strings.Contains(code, "func init()") {
		t.Fatal("expected no func init() in generated code")
	}
}

func TestGeneratedRootCaseChdirsToDoctestRoot(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "README.md", "# tree")
	writeTreeFile(t, root, "SETUP.md", setupDoc(`
type Request struct{}
type Response struct{}
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{}, nil }
`))
	writeTreeFile(t, root, "ASSERT.md", assertDoc(`
func Assert(t *testing.T, req *Request, resp *Response, err error) {}
`))

	genDir := filepath.Join(t.TempDir(), "generated")
	if err := build.Build(root, core.Options{GenDir: genDir}); err != nil {
		t.Fatalf("build: %v", err)
	}

	rootTestData, err := os.ReadFile(filepath.Join(genDir, "root_test.go"))
	if err != nil {
		t.Fatalf("read root_test.go: %v", err)
	}
	code := string(rootTestData)

	absRoot, _ := filepath.Abs(root)
	if !strings.Contains(code, "const DOCTEST_ROOT = `"+absRoot+"`") {
		t.Fatalf("expected DOCTEST_ROOT const with path %q, got:\n%s", absRoot, code)
	}
	if !strings.Contains(code, "os.Chdir(DOCTEST_ROOT)") && !strings.Contains(code, "os.Chdir(DOCTEST_ROOT);") {
		t.Fatalf("expected os.Chdir(DOCTEST_ROOT) for root-level case, got:\n%s", code)
	}
	if strings.Contains(code, "func init()") {
		t.Fatal("expected no func init() in generated code")
	}
}

func TestBuildPrintsTmpDirOnInvalidTree(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "SETUP.md", "# Setup\n\n```go\nfunc Setup(t *testing.T, req *Request) error { _ = req; return nil }\n```\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n\n```go\nfunc Check(t *testing.T, req *Request, resp *Response, err error) {}\n```\n")

	var stderr bytes.Buffer
	err := build.Build(root, core.Options{Stderr: &stderr})
	if err == nil {
		t.Fatal("expected validation error")
	}
	out := stderr.String()
	if !strings.Contains(out, "→ ") {
		t.Fatalf("expected tmp dir marker, got %q", out)
	}
}

func TestTestCommandPrintsTmpDirOnInvalidTree(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "SETUP.md", "# Setup\n\n```go\nfunc Setup(t *testing.T, req *Request) error { _ = req; return nil }\n```\n")
	writeTreeFile(t, root, "leaf/ASSERT.md", "# Assert\n\n```go\nfunc Check(t *testing.T, req *Request, resp *Response, err error) {}\n```\n")

	var stderr bytes.Buffer
	err := TestTree(root, core.Options{Stderr: &stderr})
	if err == nil {
		t.Fatal("expected validation error")
	}
	out := stderr.String()
	if !strings.Contains(out, "→ ") {
		t.Fatalf("expected tmp dir marker, got %q", out)
	}
}
