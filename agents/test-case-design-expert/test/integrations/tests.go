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
	if _, err := os.Stat(filepath.Join(wd, "main.go")); err != nil {
		fmt.Fprintf(os.Stderr, "tests must be run from the test-case-design-expert agent root (main.go not found)\n")
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

func buildAddPendingQuestions() (binPath string) {
	expertBin := buildExpert()

	tmpDir, err := os.MkdirTemp("", "test-add-pending-questions-*")
	if err != nil {
		fatal("create temp dir: %v", err)
	}
	binPath = filepath.Join(tmpDir, "add-pending-questions")

	if out, err := exec.Command("cp", expertBin, binPath).CombinedOutput(); err != nil {
		fatal("cp expert -> add-pending-questions failed: %v\n%s", err, string(out))
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

func TestAddPendingQuestionsBuild(t *T) {
	bin := buildAddPendingQuestions()
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("built binary is empty")
	}
}

func TestAddPendingQuestionsWritesQuestion(t *T) {
	bin := buildAddPendingQuestions()

	dir, err := os.MkdirTemp("", "apq-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	questionsFile := filepath.Join(dir, "questions.jsonl")

	cmd := exec.Command(bin, `{"question":"What payment methods?"}`)
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"QUESTIONS_FILE="+questionsFile,
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start add-pending-questions: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	qf, err := os.Open(qFifo)
	if err != nil {
		t.Fatalf("open question fifo for reading: %v", err)
	}
	scanner := bufio.NewScanner(qf)
	if !scanner.Scan() {
		t.Fatalf("question not written to FIFO")
	}
	line := scanner.Text()
	qf.Close()

	if !strings.Contains(line, `"type":"question"`) {
		t.Errorf("expected type:question in FIFO line, got: %s", line)
	}
	if !strings.Contains(line, "What payment methods?") {
		t.Errorf("expected question text in FIFO line, got: %s", line)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("add-pending-questions exited with error: %v\nstderr: %s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("add-pending-questions timed out")
	}

	out := stdout.String()
	if !strings.Contains(out, "questions recorded") {
		t.Errorf("expected 'questions recorded' in output, got: %s", out)
	}
	if !strings.Contains(out, "suspend") {
		t.Errorf("expected 'suspend' in output, got: %s", out)
	}
}

func TestAddPendingQuestionsMultipleQuestions(t *T) {
	bin := buildAddPendingQuestions()

	dir, err := os.MkdirTemp("", "apq-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	questionsFile := filepath.Join(dir, "questions.jsonl")

	cmd := exec.Command(bin,
		`{"question":"Q1","options":[{"label":"A"},{"label":"B"}]}`,
		`{"question":"Q2"}`,
	)
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"QUESTIONS_FILE="+questionsFile,
	)

	var stdout strings.Builder
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		t.Fatalf("start add-pending-questions: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	qf, err := os.Open(qFifo)
	if err != nil {
		t.Fatalf("open question fifo for reading: %v", err)
	}
	scanner := bufio.NewScanner(qf)
	count := 0
	for scanner.Scan() {
		count++
	}
	qf.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("add-pending-questions exited with error: %v", err)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("add-pending-questions timed out")
	}

	if count != 2 {
		t.Errorf("expected 2 questions on FIFO, got %d", count)
	}

	out := stdout.String()
	if !strings.Contains(out, "2 questions recorded") {
		t.Errorf("expected '2 questions recorded', got: %s", out)
	}
}

func TestAddPendingQuestionsWritesJSONL(t *T) {
	bin := buildAddPendingQuestions()

	dir, err := os.MkdirTemp("", "apq-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	questionsFile := filepath.Join(dir, "questions.jsonl")

	cmd := exec.Command(bin,
		`{"question":"Q1","options":[{"label":"A"}]}`,
		`{"question":"Q2"}`,
	)
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
		"QUESTIONS_FILE="+questionsFile,
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start add-pending-questions: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	go func() {
		f, _ := os.Open(qFifo)
		ioReader := bufio.NewScanner(f)
		for ioReader.Scan() {
		}
		f.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("add-pending-questions exited with error: %v", err)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("add-pending-questions timed out")
	}

	data, err := os.ReadFile(questionsFile)
	if err != nil {
		t.Fatalf("read questions.jsonl: %v", err)
	}
	content := string(data)
	if strings.Count(content, "\n") < 2 {
		t.Errorf("expected at least 2 lines in questions.jsonl, got:\n%s", content)
	}
	if !strings.Contains(content, `"type":"question"`) {
		t.Errorf("expected type:question in JSONL, got: %s", content)
	}
	if !strings.Contains(content, `"Q1"`) {
		t.Errorf("expected Q1 in JSONL, got: %s", content)
	}
	if !strings.Contains(content, `"Q2"`) {
		t.Errorf("expected Q2 in JSONL, got: %s", content)
	}
}

func TestAddPendingQuestionsNonBlocking(t *T) {
	bin := buildAddPendingQuestions()

	dir, err := os.MkdirTemp("", "apq-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	qFifo := filepath.Join(dir, "question.fifo")
	if err := syscall.Mkfifo(qFifo, 0666); err != nil {
		t.Fatalf("create question fifo: %v", err)
	}

	cmd := exec.Command(bin, `{"question":"Q1"}`)
	cmd.Env = append(os.Environ(),
		"QUESTION_FIFO="+qFifo,
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start add-pending-questions: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	start := time.Now()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	go func() {
		f, _ := os.Open(qFifo)
		ioReader := bufio.NewScanner(f)
		for ioReader.Scan() {
		}
		f.Close()
	}()

	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("add-pending-questions failed: %v", err)
		}
		if elapsed > 2*time.Second {
			t.Errorf("add-pending-questions took %v, expected near-instant (non-blocking)", elapsed)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("add-pending-questions timed out")
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
			var entry struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			}
			fmt.Sscanf(line, `{"type":"%s","id":"%s"`, &entry.Type, &entry.ID)
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
			"--model", "opencode/deepseek-v4-pro",
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
