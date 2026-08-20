package agenttty

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForOpenComposer_mcpStartingDoesNotBlock(t *testing.T) {
	orig := bannerSnapshotText
	t.Cleanup(func() { bannerSnapshotText = orig })

	const mcp = "OpenAI Codex\n• Starting MCP servers (0/8): slow_30\n› prompt\n"
	bannerSnapshotText = func(_, _ string) (string, error) {
		return mcp, nil
	}

	start := time.Now()
	waitForOpenComposer(context.Background(), "127.0.0.1:1", "s", "codex", 2*time.Second, nil)
	elapsed := time.Since(start)
	if elapsed > 400*time.Millisecond {
		t.Fatalf("--open composer wait must not hold for MCP chrome, elapsed=%v", elapsed)
	}
}

func TestWaitForOpenComposer_timeoutWhenEmpty(t *testing.T) {
	orig := bannerSnapshotText
	t.Cleanup(func() { bannerSnapshotText = orig })
	bannerSnapshotText = func(_, _ string) (string, error) {
		return "", nil
	}
	start := time.Now()
	waitForOpenComposer(context.Background(), "127.0.0.1:1", "s", "codex", 120*time.Millisecond, nil)
	if time.Since(start) < 80*time.Millisecond {
		t.Fatalf("empty snapshot should burn the timeout")
	}
}

func TestWaitForOpenComposer_abortStops(t *testing.T) {
	orig := bannerSnapshotText
	t.Cleanup(func() { bannerSnapshotText = orig })
	var n atomic.Int32
	bannerSnapshotText = func(_, _ string) (string, error) {
		n.Add(1)
		return "loading", nil
	}
	waitForOpenComposer(context.Background(), "127.0.0.1:1", "s", "codex", time.Second, func() bool { return true })
	if n.Load() > 5 {
		t.Fatalf("abort should stop quickly, polls=%d", n.Load())
	}
}
