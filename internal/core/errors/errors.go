package errors

import "fmt"

type MoreOpsError struct {
	Code    string
	Message string
	Err     error
}

func (e *MoreOpsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func Wrap(code, msg string, err error) error {
	return &MoreOpsError{Code: code, Message: msg, Err: err}
}
