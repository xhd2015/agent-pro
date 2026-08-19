// Package teststdio serializes process-global stdout/stderr redirection for
// parallel doctest leaves across agent/* packages that share one go test binary.
package teststdio

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

// mu guards os.Stdout/os.Stderr swaps process-wide. Per-package mutexes are not
// enough when agent-lib runs ./agent/subagent/... and ./agent/commit_msg/... in
// one suite — both packages race the same global streams.
var mu sync.Mutex

// Capture runs fn with os.Stdout/os.Stderr redirected to pipes and returns the
// captured output. Safe across packages under t.Parallel().
func Capture(fn func() error) (stdout, stderr string, err error) {
	mu.Lock()
	defer mu.Unlock()

	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, e := os.Pipe()
	if e != nil {
		return "", "", fmt.Errorf("stdout pipe: %w", e)
	}
	rErr, wErr, e := os.Pipe()
	if e != nil {
		_ = wOut.Close()
		_ = rOut.Close()
		return "", "", fmt.Errorf("stderr pipe: %w", e)
	}
	os.Stdout, os.Stderr = wOut, wErr

	err = fn()

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var bufOut, bufErr bytes.Buffer
	_, _ = bufOut.ReadFrom(rOut)
	_, _ = bufErr.ReadFrom(rErr)
	_ = rOut.Close()
	_ = rErr.Close()
	return bufOut.String(), bufErr.String(), err
}

// Lock acquires the process-wide stdio lock for callers that redirect streams
// themselves (must Unlock). Prefer Capture when possible.
func Lock() { mu.Lock() }

// Unlock releases the process-wide stdio lock.
func Unlock() { mu.Unlock() }
