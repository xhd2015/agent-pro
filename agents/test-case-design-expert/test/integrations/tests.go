package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var agentRoot string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(filepath.Join(wd, "ask_user")); err != nil {
		fmt.Fprintf(os.Stderr, "tests must be run from the test-case-design-expert agent root (ask_user/ directory not found)\n")
		os.Exit(1)
	}
	agentRoot = wd
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	panic(&fatalPanic{})
}

func buildExpert() (binPath string) {
	srcDir := agentRoot
	if _, err := os.Stat(filepath.Join(srcDir, "main.go")); err != nil {
		fatal("main.go not found at %s: %v — not yet implemented", srcDir, err)
	}

	tmp, err := os.CreateTemp("", "test-case-design-expert-*")
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

func buildAskUser() (binPath string) {
	expertBin := buildExpert()

	tmpDir, err := os.MkdirTemp("", "test-ask-user-*")
	if err != nil {
		fatal("create temp dir: %v", err)
	}
	binPath = filepath.Join(tmpDir, "ask_user")

	if out, err := exec.Command("cp", expertBin, binPath).CombinedOutput(); err != nil {
		fatal("cp expert -> ask_user failed: %v\n%s", err, string(out))
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		fatal("chmod: %v", err)
	}
	return binPath
}

func runExpert(args ...string) (stdout, stderr string, exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*fatalPanic); ok {
				exitCode = -1
				stderr = "build failed: source not yet implemented"
			} else {
				panic(r)
			}
		}
	}()

	bin := buildExpert()
	os.Remove(filepath.Join(agentRoot, ".session.jsonl"))
	cmd := exec.Command(bin, args...)
	cmd.Dir = agentRoot
	cmd.Stdin = nil

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

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

// ---- Tests ----

func TestAskUserBuild(t *T) {
	bin := buildAskUser()
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("built binary is empty")
	}
}

func TestAskUserWritesQuestion(t *T) {
	bin := buildAskUser()

	dir, err := os.MkdirTemp("", "ask-user-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	answerDir := filepath.Join(dir, "answer")
	os.Mkdir(answerDir, 0755)
	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	cmd := exec.Command(bin, "What payment methods are supported?")
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"ANSWER_DIR="+answerDir,
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start ask_user: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	qf, err := os.Open(qFifo)
	if err != nil {
		t.Fatalf("open question fifo for reading: %v", err)
	}
	scanner := bufio.NewScanner(qf)
	if !scanner.Scan() {
		t.Fatalf("question not written")
	}
	line := scanner.Text()
	qf.Close()

	parts := strings.SplitN(line, "\t", 2)
	if len(parts) != 2 || parts[1] != "What payment methods are supported?" {
		t.Errorf("question mismatch: got %q", line)
	}

	answerFifo := filepath.Join(answerDir, parts[0]+".fifo")
	go func() {
		f, err := os.OpenFile(answerFifo, os.O_WRONLY, 0)
		if err != nil {
			return
		}
		f.Write([]byte("credit card, PayPal"))
		f.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ask_user exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("ask_user timed out waiting for answer")
	}
}

func TestAskUserRoundtrip(t *T) {
	bin := buildAskUser()

	dir, err := os.MkdirTemp("", "ask-user-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	answerDir := filepath.Join(dir, "answer")
	os.Mkdir(answerDir, 0755)
	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	cmd := exec.Command(bin, "What analytics should the dashboard show?")
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"ANSWER_DIR="+answerDir,
	)

	var stdout strings.Builder
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		t.Fatalf("start ask_user: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	qf, err := os.Open(qFifo)
	if err != nil {
		t.Fatalf("open question fifo: %v", err)
	}
	scanner := bufio.NewScanner(qf)
	if !scanner.Scan() {
		t.Fatalf("question not written")
	}
	line := scanner.Text()
	qf.Close()

	parts := strings.SplitN(line, "\t", 2)
	if len(parts) != 2 || parts[1] != "What analytics should the dashboard show?" {
		t.Errorf("question mismatch: got %q", line)
	}

	answerFifo := filepath.Join(answerDir, parts[0]+".fifo")
	start := time.Now()
	go func() {
		f, err := os.OpenFile(answerFifo, os.O_WRONLY, 0)
		if err != nil {
			return
		}
		f.Write([]byte("page views, bounce rate"))
		f.Close()
	}()

	select {
	case err := <-done:
		elapsed := time.Since(start)
		if elapsed > 5*time.Second {
			t.Errorf("roundtrip took %v, expected <5s", elapsed)
		}
		if err != nil {
			t.Fatalf("ask_user failed: %v", err)
		}
		output := stdout.String()
		if output != "page views, bounce rate" {
			t.Errorf("stdout mismatch: got %q, want %q", output, "page views, bounce rate")
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("ask_user timed out")
	}
}

func TestMainBuild(t *T) {
	bin := buildExpert()
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("built binary is empty")
	}
}

func TestNoArgs(t *T) {
	_, stderr, code := runExpert()
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Usage") {
		t.Errorf("expected stderr to contain 'Usage', got:\n%s", stderr)
	}
}

func TestRunSelfContained(t *T) {
	if shortMode {
		t.Skip("skipping LLM-backed test in short mode")
		return
	}
	done := make(chan struct {
		out  string
		err  string
		code int
	}, 1)
	go func() {
		stdout, stderr, code := runExpert("a password reset flow: user enters email, clicks reset link, sets new password")
		done <- struct {
			out  string
			err  string
			code int
		}{stdout, stderr, code}
	}()

	select {
	case result := <-done:
		if result.code != 0 {
			t.Fatalf("binary exited %d\nstderr:\n%s", result.code, result.err)
		}
		keywords := []string{"Scenario", "Steps", "Expected", "password", "reset", "email"}
		for _, kw := range keywords {
			if !strings.Contains(result.out, kw) {
				t.Errorf("output missing keyword %q\noutput:\n%s", kw, result.out)
			}
		}
	case <-time.After(5 * time.Minute):
		t.Fatalf("binary timed out after 5m")
	}
}

func TestRunWithQuestions(t *T) {
	if shortMode {
		t.Skip("skipping LLM-backed interactive test in short mode")
		return
	}
	bin := buildExpert()

	dir, err := os.MkdirTemp("", "test-expert-interactive-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	answerDir := filepath.Join(dir, "answer")
	os.Mkdir(answerDir, 0755)
	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	cmd := exec.Command(bin, "a user dashboard that shows analytics")
	cmd.Dir = agentRoot
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"ANSWER_DIR="+answerDir,
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start expert: %v", err)
	}

	go func() {
		f, err := os.Open(qFifo)
		if err != nil {
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				id := parts[0]
				af := filepath.Join(answerDir, id+".fifo")
				time.Sleep(100 * time.Millisecond)
				wf, err := os.OpenFile(af, os.O_WRONLY, 0)
				if err == nil {
					wf.Write([]byte("page views, bounce rate, session duration"))
					wf.Close()
				}
			}
		}
	}()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		out := stdout.String()
		errOut := stderr.String()
		if err != nil {
			t.Fatalf("binary failed: %v\nstderr:\n%s", err, errOut)
		}
		keywords := []string{"Scenario", "Steps", "Expected"}
		for _, kw := range keywords {
			if !strings.Contains(out, kw) {
				t.Errorf("output missing keyword %q\noutput:\n%s", kw, out)
			}
		}
		if !strings.Contains(out, "page views") {
			t.Errorf("output missing answer content 'page views'\noutput:\n%s", out)
		}
	case <-time.After(5 * time.Minute):
		cmd.Process.Kill()
		t.Fatalf("binary timed out after 5m")
	}
}

func TestModelFlag(t *T) {
	if shortMode {
		t.Skip("skipping LLM-backed test in short mode")
		return
	}
	done := make(chan struct {
		out  string
		err  string
		code int
	}, 1)
	go func() {
		stdout, stderr, code := runExpert(
			"--model", "opencode/deepseek-v4-flash-free",
			"a simple calculator API with add, subtract, multiply, divide endpoints",
		)
		done <- struct {
			out  string
			err  string
			code int
		}{stdout, stderr, code}
	}()

	select {
	case result := <-done:
		if result.code != 0 {
			t.Fatalf("binary with --model exited %d\nstderr:\n%s", result.code, result.err)
		}
		if len(result.out) < 100 {
			t.Errorf("output too short (<100 chars), got %d chars\noutput:\n%s", len(result.out), result.out)
		}
	case <-time.After(5 * time.Minute):
		t.Fatalf("binary with --model timed out after 5m")
	}
}


