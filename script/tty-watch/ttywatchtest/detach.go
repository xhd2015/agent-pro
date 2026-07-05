package ttywatchtest

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func buildDetachRunArgv(req *Request, commandArgv []string) []string {
	argv := []string{"run", "--detach"}
	if req.CustomSessionID != "" {
		argv = append(argv, "--session-id", req.CustomSessionID)
	}
	return append(argv, commandArgv...)
}

type detachRunResult struct {
	sessionID string
	stdout    string
	stderr    string
	exitCode  int
	elapsed   time.Duration
}

func runDetachCLI(t *testing.T, req *Request, commandArgv []string) detachRunResult {
	t.Helper()
	if len(commandArgv) == 0 {
		commandArgv = []string{"sleep", "120"}
	}
	started := time.Now()
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, buildDetachRunArgv(req, commandArgv))
	if err != nil {
		t.Fatalf("detach run: %v", err)
	}
	elapsed := time.Since(started)

	firstLine := stdout
	if idx := strings.Index(stdout, "\n"); idx >= 0 {
		firstLine = stdout[:idx+1]
	}
	sessionID, ok := parseHeadlessSessionIDLine(firstLine)
	if ok && sessionID != "" {
		sid := sessionID
		t.Cleanup(func() {
			_, _, _, _ = runCLISeparate(req.Bin, req.TTYWatchHome, []string{"kill", sid})
		})
	}

	return detachRunResult{
		sessionID: sessionID,
		stdout:    stdout,
		stderr:    stderr,
		exitCode:  code,
		elapsed:   elapsed,
	}
}

func detachResponse(dr detachRunResult, home string, extra ...func(*Response)) (*Response, error) {
	ids, _ := ListRegistryIDs(home)
	resp := &Response{
		ExitCode:       dr.exitCode,
		Stdout:         dr.stdout,
		Stderr:         dr.stderr,
		Combined:       strings.TrimRight(dr.stdout+dr.stderr, "\n"),
		SessionID:      dr.sessionID,
		RegistryExists: dr.sessionID != "" && RegistryExists(home, dr.sessionID),
		RegistryIDs:    ids,
		Elapsed:        dr.elapsed,
	}
	for _, fn := range extra {
		if fn != nil {
			fn(resp)
		}
	}
	return resp, nil
}

// PhaseRunDetachPrintsSessionID runs detach and expects prompt exit 0 with session-id on stdout.
func PhaseRunDetachPrintsSessionID(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	dr := runDetachCLI(t, req, argv)
	return detachResponse(dr, req.TTYWatchHome)
}

// PhaseRunDetachNoAttachOutput runs detach without writer attach; child output stays off host stdout.
func PhaseRunDetachNoAttachOutput(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "echo DETACH_MARKER; sleep 60"}
	}
	dr := runDetachCLI(t, req, argv)
	return detachResponse(dr, req.TTYWatchHome)
}

// PhaseRunDetachSessionSurvivesInList runs detach then probes list while serve child remains live.
func PhaseRunDetachSessionSurvivesInList(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	dr := runDetachCLI(t, req, argv)
	if dr.exitCode != 0 {
		return nil, fmt.Errorf("detach exit %d stderr=%q", dr.exitCode, dr.stderr)
	}
	if dr.sessionID == "" {
		return nil, fmt.Errorf("detach missing session id in stdout %q", dr.stdout)
	}

	listOut, _, listCode, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}

	return detachResponse(dr, req.TTYWatchHome, func(resp *Response) {
		resp.ListOutput = listOut
		resp.ExitCode = listCode
		resp.SessionRunning = SessionReachable(req.TTYWatchHome, dr.sessionID)
	})
}

// PhaseRunDetachWithCustomSessionID runs detach with an explicit --session-id.
func PhaseRunDetachWithCustomSessionID(t *testing.T, req *Request) (*Response, error) {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "my-job"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	dr := runDetachCLI(t, req, argv)
	return detachResponse(dr, req.TTYWatchHome)
}

// PhaseRunDetachMutuallyExclusiveWithHeadless rejects --detach together with --headless.
func PhaseRunDetachMutuallyExclusiveWithHeadless(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"true"}
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, append([]string{"run", "--detach", "--headless"}, argv...))
	if err != nil {
		return nil, err
	}
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	return &Response{
		ExitCode:    code,
		Stdout:      stdout,
		Stderr:      stderr,
		Combined:    strings.TrimRight(stdout+stderr, "\n"),
		RegistryIDs: ids,
	}, nil
}

// PhaseRunDetachStopOnFirstArg runs detach with a command token that looks like a flag; probes via watch.
func PhaseRunDetachStopOnFirstArg(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "echo --not-a-flag"}
	}
	dr := runDetachCLI(t, req, argv)
	if dr.exitCode != 0 {
		return nil, fmt.Errorf("detach exit %d stderr=%q", dr.exitCode, dr.stderr)
	}
	if dr.sessionID == "" {
		return nil, fmt.Errorf("detach missing session id in stdout %q", dr.stdout)
	}

	watchOut, _, watchCode, watchErr := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"watch", dr.sessionID})
	if watchErr != nil {
		return nil, watchErr
	}

	return detachResponse(dr, req.TTYWatchHome, func(resp *Response) {
		resp.WatchOutput = watchOut
		resp.AltExitCode = watchCode
	})
}