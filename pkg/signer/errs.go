// pkg/signer/errors.go
package signer

import (
	"errors"
	"fmt"
)

// ErrInvalidSignature — базовая ошибка, если подпись недействительна.
var ErrInvalidSignature = errors.New("invalid signature")

// ErrBodyRead возникает при ошибке чтения тела HTTP-запроса или ответа
// во время проверки подписи. Содержит информацию о запросе и исходную ошибку.
type ErrBodyRead struct {
	Source   string // "request" или "response"
	Method   string // HTTP-метод (для ответа — пустая строка)
	URL      string // URL запроса (для ответа — пустая строка)
	BodySize int    // количество байт, которые удалось прочитать до ошибки
	Err      error  // исходная ошибка ввода-вывода
}

// Error возвращает текстовое описание ошибки.
func (e *ErrBodyRead) Error() string {
	return fmt.Sprintf("failed to read %s body: method=%s, url=%s, body_size=%d: %v",
		e.Source, e.Method, e.URL, e.BodySize, e.Err)
}

// Unwrap возвращает исходную ошибку ввода-вывода.
func (e *ErrBodyRead) Unwrap() error {
	return e.Err
}

// ErrSignedBodyMismatch — подпись тела не совпала с ожидаемой.
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

// ErrInvalidHexFormat — заголовок HashSHA256 содержит невалидный hex.
type ErrInvalidHexFormat struct {
	HeaderValue string
	Err         error
}

func (e *ErrInvalidHexFormat) Error() string {
	return fmt.Sprintf("invalid signature hex format: value=%s: %v", e.HeaderValue, e.Err)
}

func (e *ErrInvalidHexFormat) Unwrap() error {
	return e.Err
}
