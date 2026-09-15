package view

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

// testRecord is one snapshot the dashboard can chart.
func testRecord(provider usage.ProviderID, at time.Time, values map[string]float64, sessions int) usage.Record {
	return usage.Record{
		Schema:   usage.SchemaID,
		TS:       at.UTC(),
		Provider: provider,
		Usage: usage.UsageBlock{
			OK:      true,
			Values:  values,
			Display: map[string]string{"usage": "test"},
		},
		Sessions: usage.SessionsBlock{Total: sessions},
	}
}

// testStore plants records into a fresh store root.
func testStore(t *testing.T, records ...usage.Record) *usage.Store {
	t.Helper()
	store := usage.NewStore(t.TempDir())
	for _, rec := range records {
		if _, err := store.Append(rec); err != nil {
			t.Fatalf("append %s snapshot: %v", rec.Provider, err)
		}
	}
	return store
}

// freePort returns a port that was free a moment ago, used to make a port busy.
func freePort(t *testing.T) (net.Listener, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on an ephemeral port: %v", err)
	}
	return listener, listener.Addr().(*net.TCPAddr).Port
}

func TestListenRangeSkipsBusyPort(t *testing.T) {
	busy, port := freePort(t)
	defer busy.Close()

	listener, err := listenRange(port, 5)
	if err != nil {
		t.Fatalf("listenRange(%d): %v", port, err)
	}
	defer listener.Close()

	got := listener.Addr().(*net.TCPAddr).Port
	if got <= port || got >= port+5 {
		t.Fatalf("bound port = %d, want the next free port in %d..%d", got, port+1, port+4)
	}
}

func TestListenRangeEphemeral(t *testing.T) {
	listener, err := listenLocalPreferred(0)
	if err != nil {
		t.Fatalf("listenLocalPreferred(0): %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	if port <= 0 {
		t.Fatalf("bound port = %d, want a free ephemeral port", port)
	}
}

func TestListenRangeExhausted(t *testing.T) {
	first, port := freePort(t)
	defer first.Close()
	second, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port+1))
	if err != nil {
		t.Skipf("cannot occupy port %d: %v", port+1, err)
	}
	defer second.Close()

	_, err = listenRange(port, 2)
	if err == nil {
		t.Fatal("listenRange found a free port where both were busy")
	}
	if want := fmt.Sprintf("no free port in %d..%d", port, port+1); !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want it to mention %q", err, want)
	}
}

func TestParseRange(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "90m", want: 90 * time.Minute},
		{in: "24h", want: 24 * time.Hour},
		{in: "7d", want: 7 * 24 * time.Hour},
		{in: "2w", want: 14 * 24 * time.Hour},
		{in: "", wantErr: true},
		{in: "nonsense", wantErr: true},
		{in: "0d", wantErr: true},
		{in: "-1h", wantErr: true},
	}
	for _, tc := range cases {
		got, err := ParseRange(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseRange(%q) = %s, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseRange(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseRange(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

// syncBuffer is an io.Writer the server can write to from its own goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestServePrintsURLAndStops starts the real server on an ephemeral port, uses
// it over HTTP, and checks a cancelled context shuts it down.
func TestServePrintsURLAndStops(t *testing.T) {
	store := testStore(t, testRecord(usage.Grok, time.Now().Add(-time.Hour), map[string]float64{"used_percent": 42}, 7))

	stdout := &syncBuffer{}
	stderr := &syncBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, store, Options{Port: 0, Stdout: stdout, Stderr: stderr})
	}()

	url := waitForURL(t, stdout)
	body := getBody(t, url+"/api/healthz")
	if body != "ok" {
		t.Fatalf("healthz body = %q, want ok", body)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after its context was cancelled")
	}

	output := stdout.String()
	if !strings.Contains(output, "serving "+url) {
		t.Fatalf("stdout = %q, want the bound URL", output)
	}
	if !strings.Contains(output, "read-only, 1 provider, 1 snapshot") {
		t.Fatalf("stdout = %q, want the store summary line", output)
	}
	if strings.Contains(stderr.String(), "warning: port") {
		t.Fatalf("stderr = %q, want no port warning for --port 0", stderr.String())
	}
}

// waitForURL reads the served URL out of the server's stdout.
func waitForURL(t *testing.T, stdout *syncBuffer) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, line := range strings.Split(stdout.String(), "\n") {
			if strings.HasPrefix(line, "serving http://") {
				return strings.TrimPrefix(line, "serving ")
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no serving line on stdout: %q", stdout.String())
	return ""
}

func getBody(t *testing.T, url string) string {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", url, resp.StatusCode)
	}
	return string(body)
}
