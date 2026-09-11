package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/commandcode"
)

// A usage fetch that never reached the account must fail, not print an empty
// overlay with exit status 0.
func TestCommandCodeUsageFailsWhenAPIDown(t *testing.T) {
	home := writeCommandCodeHome(t)

	stdout, stderr := captureCommandCodeStreams(t, func() {
		err := handleCommandCode([]string{"usage", "--home", home, "--api-url", "http://127.0.0.1:1"})
		if err == nil {
			t.Fatalf("usage against a dead API succeeded")
		}
		if err.Error() != commandcode.ErrNetwork {
			t.Fatalf("error=%q, want %q", err, commandcode.ErrNetwork)
		}
	})

	if stdout != "" {
		t.Fatalf("stdout=%q, want no overlay", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr=%q, want the error reported only through the return value", stderr)
	}
}

func TestCommandCodeUsageFailsWhenCredentialRejected(t *testing.T) {
	home := writeCommandCodeHome(t)
	server := commandCodeFixtureServer(t,
		map[string]string{commandcode.PathWhoami: `{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"invalid api key"}}`},
		map[string]int{commandcode.PathWhoami: http.StatusUnauthorized})

	_, _ = captureCommandCodeStreams(t, func() {
		err := handleCommandCode([]string{"usage", "--home", home, "--api-url", server.URL})
		if err == nil {
			t.Fatalf("usage with a rejected credential succeeded")
		}
		if err.Error() != commandcode.ErrSessionExpired {
			t.Fatalf("error=%q, want %q", err, commandcode.ErrSessionExpired)
		}
	})
}

// A single failing endpoint degrades the overlay and stays a warning, so the
// command still succeeds.
func TestCommandCodeUsageWarnsOnPartialFailure(t *testing.T) {
	home := writeCommandCodeHome(t)
	bodies := commandCodeBillingBodies()
	bodies[commandcode.PathCredits] = `{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"invalid api key"}}`
	bodies[commandcode.PathUsageSummary] = `{"success":false,"error":{"code":"BAD_GATEWAY","status":502,"message":"upstream unavailable"}}`
	statuses := map[string]int{
		commandcode.PathCredits:      http.StatusUnauthorized,
		commandcode.PathUsageSummary: http.StatusBadGateway,
	}
	server := commandCodeFixtureServer(t, bodies, statuses)

	stdout, stderr := captureCommandCodeStreams(t, func() {
		if err := handleCommandCode([]string{"usage", "--home", home, "--api-url", server.URL}); err != nil {
			t.Fatalf("usage with a partial failure: %v", err)
		}
	})

	for _, want := range []string{" USAGE  Go Plan · active", "Plan details unavailable"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "warning:") {
		t.Fatalf("warning leaked into stdout:\n%s", stdout)
	}
	for _, want := range []string{
		"warning: " + commandcode.ErrSessionExpired,
		"warning: " + commandcode.PathUsageSummary,
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q:\n%s", want, stderr)
		}
	}
}

// Billing endpoints failing while identity succeeds is a partial failure too:
// the empty state renders and the command still succeeds.
func TestCommandCodeUsageSucceedsWithoutBillingData(t *testing.T) {
	home := writeCommandCodeHome(t)
	server := commandCodeFixtureServer(t, map[string]string{
		commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
	}, nil)

	stdout, stderr := captureCommandCodeStreams(t, func() {
		if err := handleCommandCode([]string{"usage", "--home", home, "--api-url", server.URL}); err != nil {
			t.Fatalf("usage with no billing data: %v", err)
		}
	})

	if !strings.Contains(stdout, "No billing data found.") {
		t.Fatalf("stdout missing the empty state:\n%s", stdout)
	}
	for _, path := range []string{commandcode.PathCredits, commandcode.PathSubscriptions, commandcode.PathUsageSummary} {
		if !strings.Contains(stderr, path) {
			t.Fatalf("stderr missing a warning for %s:\n%s", path, stderr)
		}
	}
}

func TestCommandCodeUsageJSONReportsPartialFailure(t *testing.T) {
	home := writeCommandCodeHome(t)
	bodies := commandCodeBillingBodies()
	bodies[commandcode.PathCredits] = `{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"nope"}}`
	server := commandCodeFixtureServer(t, bodies, map[string]int{commandcode.PathCredits: http.StatusUnauthorized})

	stdout, _ := captureCommandCodeStreams(t, func() {
		if err := handleCommandCode([]string{"usage", "--home", home, "--api-url", server.URL, "--json"}); err != nil {
			t.Fatalf("usage --json: %v", err)
		}
	})

	for _, want := range []string{`"whoami"`, `"subscription"`, `"credits": null`, `"errors"`, commandcode.ErrSessionExpired} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("json missing %q:\n%s", want, stdout)
		}
	}
}

func TestCommandCodeUsageRejectsWatchWithJSON(t *testing.T) {
	home := writeCommandCodeHome(t)

	err := handleCommandCode([]string{"usage", "--watch", "--json", "--home", home})
	if err == nil {
		t.Fatalf("--watch with --json succeeded")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error=%q, want a mutual-exclusion message", err)
	}
}

func TestCommandCodeRejectsUnknownCommand(t *testing.T) {
	err := handleCommandCode([]string{"bogus"})
	if err == nil {
		t.Fatalf("unknown command succeeded")
	}
	if !strings.Contains(err.Error(), "unknown commandcode command: bogus") {
		t.Fatalf("error=%q", err)
	}
}

func TestCommandCodeBareAndHelpPrintUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"--help"}, {"-h"}} {
		stdout, _ := captureCommandCodeStreams(t, func() {
			if err := handleCommandCode(args); err != nil {
				t.Fatalf("handleCommandCode(%v): %v", args, err)
			}
		})
		if !strings.HasPrefix(stdout, "Usage: agent-pro commandcode <command>") {
			t.Fatalf("handleCommandCode(%v) stdout:\n%s", args, stdout)
		}
	}
}

// writeCommandCodeHome creates a Command Code home holding an auth.json, the
// way `cmd login` would leave it.
func writeCommandCodeHome(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), ".commandcode")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	body := `{"apiKey":"test-key","userName":"tester","keyName":"test"}`
	if err := os.WriteFile(filepath.Join(home, commandcode.AuthFileName), []byte(body), 0o600); err != nil {
		t.Fatalf("write auth.json: %v", err)
	}
	return home
}

// commandCodeBillingBodies is a healthy individual-go payload: whoami plus an
// active subscription scoped to a period around now.
func commandCodeBillingBodies() map[string]string {
	now := time.Now()
	return map[string]string{
		commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
		commandcode.PathSubscriptions: fmt.Sprintf(
			`{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go","quantity":1,"currentPeriodStart":%q,"currentPeriodEnd":%q}}`,
			now.Add(-28*24*time.Hour).Format(time.RFC3339), now.Add(5*24*time.Hour).Format(time.RFC3339)),
		commandcode.PathCredits:      `{"credits":{"monthlyCredits":2.41},"windowLimits":{"limited":true}}`,
		commandcode.PathUsageSummary: `{"totalCount":1040,"totalCost":7.59,"successRate":100}`,
	}
}

// commandCodeFixtureServer serves bodies per path. A path absent from bodies
// answers 404 with an error envelope, which is how the API reports a missing
// billing record.
func commandCodeFixtureServer(t *testing.T, bodies map[string]string, statuses map[string]int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := bodies[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","status":404,"message":"no fixture"}}`))
			return
		}
		if status := statuses[r.URL.Path]; status != 0 {
			w.WriteHeader(status)
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

// captureCommandCodeStreams redirects stdout and stderr for the duration of fn
// and returns what each received.
func captureCommandCodeStreams(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outWriter, errWriter
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()

	fn()

	if err := outWriter.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := errWriter.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	outBytes, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	errBytes, err := io.ReadAll(errReader)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if err := outReader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	if err := errReader.Close(); err != nil {
		t.Fatalf("close stderr reader: %v", err)
	}
	return string(outBytes), string(errBytes)
}
