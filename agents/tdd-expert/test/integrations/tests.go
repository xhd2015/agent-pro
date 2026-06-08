package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	_agentRoot     string
	_agentRootOnce sync.Once
)

func resolveAgentRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed (must be run inside a git repo): %w", err)
	}
	repoRoot := strings.TrimSpace(string(out))
	agentRoot := filepath.Join(repoRoot, "agents", "tdd-expert")
	if _, err := os.Stat(filepath.Join(agentRoot, "PROMPT.md")); err != nil {
		return "", fmt.Errorf("tdd-expert agent root not found at %s: PROMPT.md missing", agentRoot)
	}
	return agentRoot, nil
}

func agentRootPath() string {
	_agentRootOnce.Do(func() {
		var err error
		_agentRoot, err = resolveAgentRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	})
	return _agentRoot
}

func fatal(format string, args ...any) {
	panic(&fatalPanic{})
}

func buildExpert() (binPath string) {
	srcDir := agentRootPath()
	if _, err := os.Stat(filepath.Join(srcDir, "main.go")); err != nil {
		fatal("main.go not found at %s: %v", srcDir, err)
	}

	tmp, err := os.CreateTemp("", "tdd-expert-*")
	if err != nil {
		fatal("create temp file: %v", err)
	}
	binPath = tmp.Name()
	tmp.Close()

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = srcDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fatal("go build expert failed: %v\n%s", err, string(out))
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		fatal("chmod: %v", err)
	}
	return binPath
}

func runExpert(bin string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = agentRootPath()
	cmd.Stdin = nil

	var outBuf, errBuf strings.Builder
	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	} else {
		exitCode = 0
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func fixturesDir() string {
	return filepath.Join(agentRootPath(), "test", "integrations")
}

func TestBuild(t *T) {
	bin := buildExpert()
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("built binary is empty")
	}
	t.Logf("binary: %s (%d bytes)", bin, info.Size())
}

func TestNoArgs(t *T) {
	bin := buildExpert()
	_, stderr, code := runExpert(bin)
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Usage") && !strings.Contains(stderr, "usage") {
		t.Errorf("expected stderr to contain usage info, got:\n%s", stderr)
	}
}

func TestHelpFlag(t *T) {
	bin := buildExpert()
	stdout, _, code := runExpert(bin, "-h")
	if code != 0 {
		t.Fatalf("expected exit 0 for -h, got %d", code)
	}
	if !strings.Contains(stdout, "Usage") {
		t.Errorf("expected -h output to contain 'Usage', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "test-case-tree-dir") {
		t.Errorf("expected -h output to mention test-case-tree-dir, got:\n%s", stdout)
	}
}

func TestWriteCLI(t *T) {
	if shortMode {
		t.Skip("skipping LLM-backed test in short mode")
		return
	}

	t.Logf("building tdd-expert binary...")
	bin := buildExpert()
	t.Logf("built: %s", bin)

	testCaseDir := filepath.Join(fixturesDir(), "write-a-cli", "write-cli-test-cases")
	outDir := filepath.Join(testCaseDir, "cli")

	t.Logf("removing existing output dir: %s", outDir)
	os.RemoveAll(outDir)

	t.Logf("running tdd-expert on %s", testCaseDir)
	stdout, stderr, code := runExpert(bin, testCaseDir, "-o", outDir)
	if code != 0 {
		t.Fatalf("expert exited %d\nstderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	t.Logf("output dir: %s", outDir)

	validateGeneratedModule(t, outDir, "write-cli")
}

func TestWriteLibrary(t *T) {
	if shortMode {
		t.Skip("skipping LLM-backed test in short mode")
		return
	}

	t.Logf("building tdd-expert binary...")
	bin := buildExpert()
	t.Logf("built: %s", bin)

	testCaseDir := filepath.Join(fixturesDir(), "write-a-library", "math-lib-test-cases")
	outDir := filepath.Join(fixturesDir(), "write-a-library", "lib")

	t.Logf("removing existing output dir: %s", outDir)
	os.RemoveAll(outDir)

	t.Logf("running tdd-expert on %s", testCaseDir)
	stdout, stderr, code := runExpert(bin, testCaseDir, "-o", outDir)
	if code != 0 {
		t.Fatalf("expert exited %d\nstderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	t.Logf("output dir: %s", outDir)

	validateGeneratedModule(t, outDir, "math-lib")
}

func validateGeneratedModule(t *T, dir, expectedSlug string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("output dir is empty; expected generated Go files")
	}

	hasGoMod := false
	hasStubGo := false
	hasTestGo := false
	for _, f := range files {
		name := f.Name()
		if name == "go.mod" {
			hasGoMod = true
		}
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			hasStubGo = true
		}
		if strings.HasSuffix(name, "_test.go") {
			hasTestGo = true
		}
	}
	if !hasGoMod {
		t.Errorf("output dir missing go.mod")
	}
	if !hasStubGo {
		t.Errorf("output dir missing stub .go file")
	}
	if !hasTestGo {
		t.Errorf("output dir missing _test.go file(s)")
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go build failed:\n%s", string(out))
	} else {
		t.Logf("go build passed")
	}

	cmd = exec.Command("go", "test", "./...", "-v", "-count=1")
	cmd.Dir = dir
	out, err = cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected go test to fail (RED state), but it passed. All stubs should return errors.")
	}
	output := string(out)
	if !strings.Contains(output, "not implemented") {
		t.Errorf("expected 'not implemented' in test output, got:\n%s", output)
	}
	t.Logf("go test correctly failed (RED state)")
}

func TestAgentRootResolves(t *T) {
	root, err := resolveAgentRoot()
	if err != nil {
		t.Fatalf("resolveAgentRoot failed: %v", err)
	}
	if !strings.HasSuffix(root, filepath.Join("agents", "tdd-expert")) {
		t.Fatalf("expected agent root to end with agents/tdd-expert, got: %s", root)
	}
	if _, err := os.Stat(filepath.Join(root, "PROMPT.md")); err != nil {
		t.Fatalf("PROMPT.md not found at resolved root %s: %v", root, err)
	}
	path, err := filepath.Rel(agentRootPath(), filepath.Join(agentRootPath(), "test", "integrations"))
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}
	if path != "test/integrations" {
		t.Errorf("expected fixtures to be at test/integrations relative to agent root, got: %s", path)
	}
	t.Logf("agent root: %s", root)
}
