package main

import (
	"fmt"
	"time"
)

type T struct {
	name    string
	failed  bool
	skipped bool
	start   time.Time
	elapsed time.Duration
}

func (t *T) Fail() { t.failed = true }

func (t *T) Skip(reason string) {
	t.skipped = true
}

func (t *T) Skipf(format string, args ...any) {
	t.skipped = true
	fmt.Printf("    skip: %s\n", fmt.Sprintf(format, args...))
}

func (t *T) Errorf(format string, args ...any) {
	t.failed = true
	fmt.Printf("    %s\n", fmt.Sprintf(format, args...))
}

func (t *T) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Printf("    %s\n", fmt.Sprintf(format, args...))
	panic(&fatalPanic{})
}

func (t *T) Logf(format string, args ...any) {
	fmt.Printf("    %s\n", fmt.Sprintf(format, args...))
}

type fatalPanic struct{}
