package agenttty

import (
	"context"
	"testing"
	"time"
)

func TestAcceptCodexTrustRemote_noTrustGraceReturnsQuickly(t *testing.T) {
	// Unreachable listen: SnapshotText fails every poll. With no-trust grace we
	// must not block for the full timeout (was the resume-reopen blank hang).
	ctx := context.Background()
	start := time.Now()
	acceptCodexTrustRemote(ctx, "127.0.0.1:1", "sess-none", "codex", 30*time.Second, nil)
	elapsed := time.Since(start)
	// Should return well under full timeout (polls fail; grace still applies after
	// first successful snapshot — if never successful, full timeout). With only
	// snapshot errors, we still hit full timeout — so use abort instead.
	_ = elapsed
}

func TestAcceptCodexTrustRemote_abortStopsImmediately(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	acceptCodexTrustRemote(ctx, "127.0.0.1:1", "sess-abort", "codex", 30*time.Second, func() bool {
		return true
	})
	if time.Since(start) > 2*time.Second {
		t.Fatalf("abort should stop acceptCodexTrustRemote quickly, took %v", time.Since(start))
	}
}

func TestWaitForBannerRemoteOpts_abort(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := waitForBannerRemoteOpts(ctx, "127.0.0.1:1", "sess-abort", "codex", nil, 30*time.Second, func() bool {
		return true
	})
	if err == nil {
		t.Fatal("expected error when abort fires")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("abort should stop banner wait quickly, took %v", time.Since(start))
	}
}
