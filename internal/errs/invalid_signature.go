package errs

import (
	"errors"
	"fmt"
)

var ErrInvalidSignature = errors.New("invalid signature")

// ===================================

type ErrSignedBodyMismatch struct {
	Expected string
	Received string
}

func (e *ErrSignedBodyMismatch) Error() string {
	return fmt.Sprintf("signed body mismatch: expected=%s, received=%s", e.Expected, e.Received)
}

func (e *ErrSignedBodyMismatch) Unwrap() error {
	return ErrInvalidSignature
}

// ===================================

type ErrInvalidHexFormat struct {
	HeaderValue string
	Err         error
}

func (e *ErrInvalidHexFormat) Error() string {
	return fmt.Sprintf(
		"invalid signature hex format: value=%s: %v",
		e.HeaderValue, e.Err,
	)
}

func (e *ErrInvalidHexFormat) Unwrap() []error {
	return []error{ErrInvalidSignature, e.Err}
}
