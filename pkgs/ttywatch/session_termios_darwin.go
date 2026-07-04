//go:build darwin

package ttywatch

import "golang.org/x/sys/unix"

func ioctlGetTermios() uint {
	return unix.TIOCGETA
}

func ioctlSetTermios() uint {
	return unix.TIOCSETA
}