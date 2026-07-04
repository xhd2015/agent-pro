//go:build linux

package main

import "golang.org/x/sys/unix"

func ioctlGetTermios() uint {
	return unix.TIOCGETS
}

func ioctlSetTermios() uint {
	return unix.TIOCSETS
}