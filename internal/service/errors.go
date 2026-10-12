package service

import "errors"

// ValidationError signals a client-input problem that should surface as a
// 400 response (as opposed to a not-found or infrastructure error).
type ValidationError struct {
	Text string
}

func (e *ValidationError) Error() string {
	return e.Text
}

// ErrNotFound is returned when a book id does not exist. Repositories return
// it, and callers use it to tell not-found apart from other errors.
var ErrNotFound = errors.New("book not found")
