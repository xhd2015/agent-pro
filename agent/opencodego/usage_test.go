package opencodego

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFetchReadsTokenFileAndMeters(t *testing.T) {
	var gotAuth, gotUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		if r.URL.Path != PathUsage {
			t.Errorf("path = %s, want %s", r.URL.Path, PathUsage)
		}
		_, _ = w.Write([]byte(`{"usage":{"rolling":{"status":"ok","percent":18,"resetsAt":"2026-11-23T15:00:00.000Z"},"weekly":{"status":"ok","percent":41,"resetsAt":"2026-11-30T00:00:00.000Z"},"monthly":{"status":"ok","percent":32,"resetsAt":"2026-12-22T13:51:41.000Z"}}}`))
	}))
	t.Cleanup(server.Close)

	tokenPath := filepath.Join(t.TempDir(), "token.txt")
	if err := os.WriteFile(tokenPath, []byte(" oc_sk_test \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	usage, err := Fetch(context.Background(), FetchOpts{TokenFile: tokenPath, BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer oc_sk_test" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotUA != UserAgent {
		t.Fatalf("User-Agent = %q, want %q", gotUA, UserAgent)
	}
	if usage.Monthly.Percent != 32 || usage.Weekly.Percent != 41 || usage.Rolling.Percent != 18 {
		t.Fatalf("meters = %+v", usage)
	}
	monthly := usage.Monthly.ResetTime()
	if monthly.IsZero() || !monthly.Equal(time.Date(2026, 12, 22, 13, 51, 41, 0, time.UTC)) {
		t.Fatalf("monthly reset = %s", monthly)
	}
}

func TestFetchReportsEntitlementError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"EntitlementError","message":"OpenCode Go subscription required."}}`))
	}))
	t.Cleanup(server.Close)
	_, err := Fetch(context.Background(), FetchOpts{Token: "oc_sk_x", BaseURL: server.URL})
	if err == nil || !strings.Contains(err.Error(), "OpenCode Go subscription required") || !strings.Contains(err.Error(), "403") {
		t.Fatalf("error = %v", err)
	}
}

func TestFetchMissingTokenFile(t *testing.T) {
	_, err := Fetch(context.Background(), FetchOpts{TokenFile: filepath.Join(t.TempDir(), "missing.txt")})
	if err == nil || !strings.Contains(err.Error(), "read OpenCode token file") {
		t.Fatalf("error = %v", err)
	}
}
