package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

func main() {
	// Resolve binaries
	mockBinary := "/tmp/llm-mock-run-commandcode"
	if len(os.Args) >= 2 {
		mockBinary = os.Args[1]
	}
	absMock, err := filepath.Abs(mockBinary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve mock binary: %v\n", err)
		os.Exit(1)
	}

	ttyWatchBin := "/tmp/tty-watch-inspect"
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", ttyWatchBin, "./script/tty-watch")
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: build tty-watch: %v\n", err)
		os.Exit(1)
	}

	outDir := filepath.Join("script", "debug", "cmd-mock-inspect", "out")
	os.MkdirAll(outDir, 0755)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	sessionID := "inspect-interactive"
	session := ttywatch.NewEphemeralSession(
		ttywatch.DefaultRegistryConfig("").Home,
		sessionID,
		[]string{absMock},
	)
	defer session.Kill()

	fmt.Println("CHECK: start interactive cmd session in PTY")
	if err := session.StartDetached(ctx, ttyWatchBin); err != nil {
		fmt.Printf("FAIL: start session: %v\n", err)
		os.Exit(1)
	}

	// Wait for cmd TUI to render
	fmt.Println("CHECK: wait for cmd TUI to appear")
	time.Sleep(5 * time.Second)

	// Send "hello" and Enter
	fmt.Println("CHECK: send 'hello' + Enter to interactive cmd")
	if err := session.Send("hello"); err != nil {
		fmt.Printf("FAIL: send hello: %v\n", err)
		os.Exit(1)
	}
	time.Sleep(500 * time.Millisecond)
	if err := session.Send("\r"); err != nil {
		fmt.Printf("FAIL: send enter: %v\n", err)
		os.Exit(1)
	}

	// Wait for cmd to process
	fmt.Println("CHECK: wait for cmd to process and respond")
	time.Sleep(15 * time.Second)

	// Take snapshot
	snapshot, err := session.Snapshot()
	if err != nil {
		fmt.Printf("FAIL: snapshot: %v\n", err)
		os.Exit(1)
	}

	// Save snapshot
	snapFile := filepath.Join(outDir, "snapshot.txt")
	os.WriteFile(snapFile, []byte(snapshot), 0644)

	fmt.Println()
	fmt.Println("EVIDENCE:", snapFile)
	fmt.Println("--- snap ---")
	fmt.Println(snapshot)
	fmt.Println("--- end ---")

	// Check for symptom
	hasRetrying := strings.Contains(snapshot, "Retrying")
	hasConnectionIssue := strings.Contains(snapshot, "Connection Issue")

	if hasRetrying || hasConnectionIssue {
		fmt.Println()
		fmt.Println("RESULT: FAIL")
		fmt.Println("REASON: scrollback contains 'Connection Issue' or 'Retrying'")
		os.Exit(1)
	}

	// Check for successful response
	hasResponse := strings.Contains(snapshot, "hello") || strings.Contains(snapshot, "Hello")

	if !hasResponse {
		fmt.Println()
		fmt.Println("RESULT: FAIL")
		fmt.Println("REASON: no 'hello' or 'Hello' response found in scrollback")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("RESULT: PASS")
	fmt.Println("REASON: interactive cmd returned mock response without retry errors")
}
