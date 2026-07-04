//go:build linux

package ttywatch

import "golang.org/x/sys/unix"

func ioctlGetTermios() uint {
	return unix.TIOCGETS
}

func ioctlSetTermios() uint {
	return unix.TIOCSETS
}