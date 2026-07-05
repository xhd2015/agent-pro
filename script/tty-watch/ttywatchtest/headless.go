package ttywatchtest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

const headlessSessionIDPrefix = "session-id: "

func buildHeadlessRunArgv(req *Request, commandArgv []string) []string {
	argv := []string{"run", "--headless"}
	if req.CustomSessionID != "" {
		argv = append(argv, "--session-id", req.CustomSessionID)
	}
	return append(argv, commandArgv...)
}

func parseHeadlessSessionIDLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, headlessSessionIDPrefix) {
		return "", false
	}
	id := strings.TrimSpace(line[len(headlessSessionIDPrefix):])
	if id == "" {
		return "", false
	}
	return id, true
}

type headlessRun struct {
	cmd        *exec.Cmd
	stdoutPipe io.ReadCloser
	stdout     *bufio.Reader
	stderrBuf  *bytes.Buffer
	sessionID  string
	firstLine  string
	started    time.Time
}

func startHeadlessRun(t *testing.T, req *Request) *headlessRun {
	t.Helper()
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	cmd := exec.Command(req.Bin, buildHeadlessRunArgv(req, argv)...)
	cmd.Env = envWithHome(req.TTYWatchHome)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("headless stdout pipe: %v", err)
	}
	stderrBuf := &bytes.Buffer{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("headless start: %v", err)
	}

	reader := bufio.NewReader(stdoutPipe)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		t.Fatalf("read headless session-id line: %v stderr=%q", err, stderrBuf.String())
	}
	sessionID, ok := parseHeadlessSessionIDLine(firstLine)
	if !ok {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		t.Fatalf("invalid headless session-id line %q stderr=%q", firstLine, stderrBuf.String())
	}

	hr := &headlessRun{
		cmd:        cmd,
		stdoutPipe: stdoutPipe,
		stdout:     reader,
		stderrBuf:  stderrBuf,
		sessionID:  sessionID,
		firstLine:  firstLine,
		started:    time.Now(),
	}
	t.Cleanup(func() { terminateHeadlessCmd(t, hr.cmd) })
	return hr
}

func terminateHeadlessCmd(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		<-done
	}
}

func drainHeadlessStdout(hr *headlessRun, timeout time.Duration) string {
	if hr == nil || hr.stdoutPipe == nil {
		return ""
	}
	type readResult struct {
		data string
	}
	ch := make(chan readResult, 1)
	go func() {
		buf, _ := io.ReadAll(hr.stdoutPipe)
		ch <- readResult{string(buf)}
	}()
	select {
	case res := <-ch:
		return res.data
	case <-time.After(timeout):
		return ""
	}
}

func waitHeadlessExit(hr *headlessRun, timeout time.Duration) (int, time.Duration, error) {
	done := make(chan error, 1)
	go func() { done <- hr.cmd.Wait() }()
	select {
	case err := <-done:
		elapsed := time.Since(hr.started)
		if err == nil {
			return 0, elapsed, nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), elapsed, nil
		}
		return -1, elapsed, err
	case <-time.After(timeout):
		return -1, time.Since(hr.started), fmt.Errorf("headless wait timeout after %s", timeout)
	}
}

func headlessResponse(hr *headlessRun, exitCode int, elapsed time.Duration, extraStdout string, home string) (*Response, error) {
	stdout := hr.firstLine + extraStdout
	stderr := hr.stderrBuf.String()
	ids, _ := ListRegistryIDs(home)
	return &Response{
		ExitCode:       exitCode,
		Stdout:         stdout,
		Stderr:         stderr,
		Combined:       strings.TrimRight(stdout+stderr, "\n"),
		SessionID:      hr.sessionID,
		RegistryExists: RegistryExists(home, hr.sessionID),
		RegistryIDs:    ids,
		Elapsed:        elapsed,
	}, nil
}

// PhaseRunHeadlessPrintsSessionID runs headless and returns the session-id line on stdout.
func PhaseRunHeadlessPrintsSessionID(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	extra := drainHeadlessStdout(hr, 400*time.Millisecond)
	resp, err := headlessResponse(hr, 0, time.Since(hr.started), extra, req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.Stdout = hr.firstLine
	return resp, nil
}

// PhaseRunHeadlessNoAttachOutput runs headless without writer attach; child output stays off host stdout.
func PhaseRunHeadlessNoAttachOutput(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "echo HEADLESS_MARKER; sleep 60"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	extra := drainHeadlessStdout(hr, 500*time.Millisecond)
	return headlessResponse(hr, 0, time.Since(hr.started), extra, req.TTYWatchHome)
}

// PhaseRunHeadlessWaitsUntilChildExits blocks until the headless child exits naturally.
func PhaseRunHeadlessWaitsUntilChildExits(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"true"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	code, elapsed, err := waitHeadlessExit(hr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	resp, err := headlessResponse(hr, code, elapsed, "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.RegistryIDs = ids
	resp.RegistryExists = len(ids) > 0
	return resp, nil
}

// PhaseRunHeadlessCtrlCForwardsExits sends SIGINT to the headless parent and waits for exit.
func PhaseRunHeadlessCtrlCForwardsExits(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "trap 'exit 0' INT; sleep 300"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	if err := hr.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return nil, fmt.Errorf("sigint headless parent: %w", err)
	}
	code, elapsed, err := waitHeadlessExit(hr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	resp, err := headlessResponse(hr, code, elapsed, "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.RegistryIDs = ids
	resp.RegistryExists = len(ids) > 0
	return resp, nil
}

// PhaseRunHeadlessCtrlCWaitingLogs sends SIGINT and captures the waiting-for-exit stderr line timing.
func PhaseRunHeadlessCtrlCWaitingLogs(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	if err := hr.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return nil, fmt.Errorf("sigint headless parent: %w", err)
	}

	const waitingLine = "waiting for program to exit..."
	exitCh := make(chan int, 1)
	go func() {
		err := hr.cmd.Wait()
		if err == nil {
			exitCh <- 0
			return
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCh <- exitErr.ExitCode()
			return
		}
		exitCh <- -1
	}()

	var waitingAfter time.Duration
	sawWaiting := false
	deadline := time.Now().Add(14 * time.Second)
	code := -1

	for time.Now().Before(deadline) {
		if !sawWaiting && strings.Contains(hr.stderrBuf.String(), waitingLine) {
			sawWaiting = true
			waitingAfter = time.Since(hr.started)
		}
		select {
		case code = <-exitCh:
			goto finished
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	select {
	case code = <-exitCh:
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("headless still running after ctrl-c grace window")
	}

finished:

	elapsed := time.Since(hr.started)
	time.Sleep(200 * time.Millisecond)
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	resp, err := headlessResponse(hr, code, elapsed, "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.RegistryIDs = ids
	resp.RegistryExists = len(ids) > 0
	if sawWaiting {
		resp.AltStderr = waitingAfter.String()
	}
	return resp, nil
}

// PhaseRunHeadlessSessionLiveWhileWaiting probes list while headless parent is still blocked.
func PhaseRunHeadlessSessionLiveWhileWaiting(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)

	listOut, _, listCode, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	if err := hr.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		return nil, fmt.Errorf("headless parent exited before list probe: %w", err)
	}

	resp, err := headlessResponse(hr, listCode, time.Since(hr.started), "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.ListOutput = listOut
	resp.ExitCode = listCode
	resp.SessionRunning = SessionReachable(req.TTYWatchHome, hr.sessionID)
	return resp, nil
}

// PhaseRunHeadlessWithCustomSessionID runs headless with an explicit --session-id.
func PhaseRunHeadlessWithCustomSessionID(t *testing.T, req *Request) (*Response, error) {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "my-job"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)
	resp, err := headlessResponse(hr, 0, time.Since(hr.started), "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.Stdout = hr.firstLine
	return resp, nil
}

// PhaseRunHeadlessStopOnFirstArg runs watch with the first program arg as session id (stop subcommand).
func PhaseRunHeadlessStopOnFirstArg(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "echo --not-a-flag"}
	}
	req.RunCommand = argv
	hr := startHeadlessRun(t, req)

	watchOut, _, watchCode, watchErr := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"watch", hr.sessionID})
	if watchErr != nil {
		return nil, watchErr
	}

	code, elapsed, err := waitHeadlessExit(hr, 8*time.Second)
	if err != nil {
		return nil, err
	}
	resp, err := headlessResponse(hr, code, elapsed, "", req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	resp.WatchOutput = watchOut
	resp.AltExitCode = watchCode
	return resp, nil
}