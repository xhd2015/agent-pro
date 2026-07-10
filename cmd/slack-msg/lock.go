package main

import (
	"fmt"
	"os"
	"syscall"
)

type processLock struct {
	file *os.File
}

func acquireLock(path string) (*processLock, error) {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another slack-msg is already running")
	}
	return &processLock{file: f}, nil
}

func (l *processLock) release() {
	if l == nil || l.file == nil {
		return
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
	l.file = nil
}

func dirOf(path string) string {
	if i := len(path) - 1; i >= 0 {
		for j := i; j >= 0; j-- {
			if path[j] == '/' {
				if j == 0 {
					return "/"
				}
				return path[:j]
			}
		}
	}
	return "."
}
