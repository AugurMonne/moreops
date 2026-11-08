//go:build !linux

package wizard

import "errors"

type rawState struct{}

func isTerminal(fd uintptr) bool {
	return false
}

func makeRaw(fd int) (*rawState, error) {
	return nil, errors.New("raw mode fallback for non-linux OS")
}

func restoreTerminal(fd int, old *rawState) {
}
