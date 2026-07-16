package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var outDir = filepath.Join("script", "debug", "cmd-mock-send-e2e", "out")

func main() {
	os.MkdirAll(outDir, 0755)

	agentRun := "/tmp/agent-run"
	mockBin := "/tmp/llm-mock-run-commandcode"

	// Build
	build := exec.Command("go", "build", "-o", agentRun, "./cmd/agent-run")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Printf("FAIL: build agent-run: %v\n%s\n", err, out)
		os.Exit(1)
	}
	build2 := exec.Command("go", "build", "-o", mockBin, "./agent/llm/llm-mock/llm-mock-run-commandcode")
	if out, err := build2.CombinedOutput(); err != nil {
		fmt.Printf("FAIL: build mock: %v\n%s\n", err, out)
		os.Exit(1)
	}
	fmt.Println("CHECK: binaries built")

	// Cleanup
	exec.Command("pkill", "-9", "-f", "__serve__.*llm-mock").Run()
	exec.Command("pkill", "-9", "-f", "llm-mock-run-commandcode").Run()
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	_ = ctx

	fmt.Println("CHECK: open session with 'Hello' prompt")
	runCmd := exec.Command(agentRun, "run",
		"--agent-runner", "commandcode-tty",
		"--agent-runner-binary", mockBin,
		"--session-id", "e2e-test",
		"--open", "Hello",
	)
	runOut, err := runCmd.CombinedOutput()
	fmt.Printf("  run output: %s\n", strings.TrimSpace(string(runOut)))
	saveArtifact("run-output.txt", string(runOut))
	if err != nil {
		fmt.Printf("FAIL: run --open: %v\n", err)
		os.Exit(1)
	}

	// Wait for PTY to process first message
	fmt.Println("CHECK: wait for first response")
	time.Sleep(10 * time.Second)

	snap1 := execSnapshot(agentRun, "e2e-test")
	saveArtifact("snapshot-after-open.txt", snap1)

	if !strings.Contains(snap1, "Hello") {
		fmt.Printf("FAIL: snapshot missing 'Hello'\nEVIDENCE:\n%s\n", snap1)
		os.Exit(1)
	}
	fmt.Println("  first response present")

	// Send second message
	fmt.Println("CHECK: send 'Hello 2'")
	sendCmd := exec.Command(agentRun, "send", "--max-wait", "15s", "e2e-test", "Hello 2")
	sendOut, sendErr := sendCmd.CombinedOutput()
	sendResult := strings.TrimSpace(string(sendOut))
	fmt.Printf("  send: %s (err=%v)\n", sendResult, sendErr)
	saveArtifact("send-output.txt", sendResult)

	// Wait for response
	fmt.Println("CHECK: wait for second response")
	deadline := time.Now().Add(25 * time.Second)
	var snap2 string
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)
		snap2 = execSnapshot(agentRun, "e2e-test")
		if snap2 != snap1 && strings.Contains(snap2, "Hello 2") {
			break
		}
	}
	saveArtifact("snapshot-after-send.txt", snap2)

	if !strings.Contains(snap2, "Hello 2") {
		fmt.Printf("FAIL: snapshot missing 'Hello 2'\nEVIDENCE:\n%s\n", snap2)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("RESULT: PASS")
	fmt.Println("REASON: open + send both delivered; scrollback contains both prompts")
}

func execSnapshot(agentRun, sessionID string) string {
	cmd := exec.Command(agentRun, "tty", "snapshot", sessionID)
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func saveArtifact(name, content string) {
	os.WriteFile(filepath.Join(outDir, name), []byte(content), 0644)
}
