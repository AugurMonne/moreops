//go:build linux

package wizard

import (
	"syscall"
	"unsafe"
)

type rawState = syscall.Termios

func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func makeRaw(fd int) (*rawState, error) {
	var oldTermios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldTermios)), 0, 0, 0)
	if err != 0 {
		return nil, err
	}

	raw := oldTermios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	_, _, err = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw)), 0, 0, 0)
	if err != 0 {
		return nil, err
	}

	return &oldTermios, nil
}

func restoreTerminal(fd int, old *rawState) {
	if old != nil {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(old)), 0, 0, 0)
	}
}
