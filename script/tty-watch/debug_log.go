package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const envDebugLogFile = "TTY_WATCH_DEBUG_LOG_FILE"

var (
	debugLogMu     sync.Mutex
	debugLogFile   *os.File
	debugLogPath   string
	debugLogOpened bool
)

func debugLogEnabled() bool {
	return os.Getenv(envDebugLogFile) != ""
}

func debugLogOpen() *os.File {
	path := os.Getenv(envDebugLogFile)
	if path == "" {
		return nil
	}
	debugLogMu.Lock()
	defer debugLogMu.Unlock()
	if debugLogOpened && debugLogPath == path && debugLogFile != nil {
		return debugLogFile
	}
	if debugLogFile != nil {
		_ = debugLogFile.Close()
		debugLogFile = nil
		debugLogOpened = false
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	debugLogFile = f
	debugLogPath = path
	debugLogOpened = true
	return debugLogFile
}

func debugLogf(format string, args ...any) {
	f := debugLogOpen()
	if f == nil {
		return
	}
	debugLogMu.Lock()
	defer debugLogMu.Unlock()
	ts := time.Now().Format(time.RFC3339Nano)
	pid := os.Getpid()
	_, _ = fmt.Fprintf(f, "%s pid=%d %s", ts, pid, fmt.Sprintf(format, args...))
	if len(format) == 0 || format[len(format)-1] != '\n' {
		_, _ = f.WriteString("\n")
	}
}

func debugLogBytes(label string, data []byte) {
	if !debugLogEnabled() {
		return
	}
	const maxDump = 512
	n := len(data)
	dump := data
	truncated := false
	if n > maxDump {
		dump = data[:maxDump]
		truncated = true
	}
	quoted := strconv.QuoteToASCII(string(dump))
	hexLine := hex.EncodeToString(dump)
	if truncated {
		debugLogf("%s len=%d truncated_to=%d quoted=%s hex=%s", label, n, maxDump, quoted, hexLine)
		return
	}
	debugLogf("%s len=%d quoted=%s hex=%s", label, n, quoted, hexLine)
}