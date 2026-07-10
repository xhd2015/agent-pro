// Command lockholder holds an exclusive flock on a path so tty-watch doctests
// can assert registry-lock-busy diagnostics against a known holder PID/command.
//
// Usage:
//
//	lockholder <lock-path> <marker>
//
// Prints the holder PID on stdout (single line), then keeps the flock until killed.
// Also spawns a long-lived child (`lockholder --child <marker>-child`) so process-tree
// diagnostics can show children of the holder.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--child" {
		// Distinctive argv: lockholder --child <marker>-child
		// Block until parent/test kills the tree.
		select {}
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: lockholder <lock-path> <marker>")
		os.Exit(2)
	}
	lockPath := os.Args[1]
	marker := os.Args[2]

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open lock: %v\n", err)
		os.Exit(1)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		fmt.Fprintf(os.Stderr, "flock: %v\n", err)
		os.Exit(1)
	}

	// Child for process-tree "children of holder" section.
	child := exec.Command(os.Args[0], "--child", marker+"-child")
	if err := child.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start child: %v\n", err)
		os.Exit(1)
	}

	// Harness reads this line before launching tty-watch.
	fmt.Printf("%d\n", os.Getpid())

	// Hold forever (test Cleanup kills us).
	for {
		time.Sleep(time.Hour)
	}
}
