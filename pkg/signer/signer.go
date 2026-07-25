// Package signer предоставляет создание и проверку HMAC-SHA256 подписей
// для HTTP-запросов и ответов. Используется для подтверждения целостности
// и подлинности передаваемых данных между клиентом и сервером.
//
// Клиент подписывает тело запроса и проверяет подпись ответа.
// Сервер проверяет подпись входящего запроса и подписывает ответ.
// Обе стороны используют общий секретный ключ.
package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"net/http"

	"go.uber.org/zap"
)

// Signer создаёт и проверяет HMAC-SHA256 подписи HTTP-запросов и ответов.
// Использует общий секретный ключ (secretKey) — один и тот же ключ должен быть
// у клиента и сервера.
type Signer struct {
	secretKey string
	logger    *zap.Logger
}

// New создаёт новый Signer.
func New(secretKey string, logger *zap.Logger) *Signer {
	if secretKey == "" {
		logger.Warn("signer initialized with empty secret key - signatures disabled")
		return nil
	}
	return &Signer{
		secretKey: secretKey,
		logger:    logger,
	}
}

// CreateSignature вычисляет HMAC-SHA256 подпись для переданных данных.
// Возвращает бинарную подпись ([]byte), которую можно закодировать в hex или base64.
func (s *Signer) CreateSignature(data []byte) []byte {
	hmacHash := hmac.New(sha256.New, []byte(s.secretKey))
	hmacHash.Write(data)
	return hmacHash.Sum(nil)
}

// CheckResponseSignature проверяет подпись HTTP-ответа.
// Читает тело ответа, вычисляет ожидаемую подпись и сверяет с заголовком HashSHA256.
// Возвращает SignatureError, если подпись не совпадает или заголовок имеет неверный формат.
func (s *Signer) CheckResponseSignature(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ErrBodyRead{Source: "response", BodySize: len(body), Err: err}
	}

	headerValue := resp.Header.Get("HashSHA256")
	received, err := hex.DecodeString(headerValue)
	if err != nil {
		return &ErrInvalidHexFormat{
			HeaderValue: headerValue,
			Err:         err,
		}
	}

	expected := s.CreateSignature(body)

	if !hmac.Equal(received, expected) {
		return &ErrSignedBodyMismatch{
			Expected: hex.EncodeToString(expected),
			Received: hex.EncodeToString(received),
		}
	}

	return nil
}

// CheckRequestSignature проверяет подпись HTTP-запроса.
// Если заголовок HashSHA256 отсутствует — проверка пропускается (нет подписи — нет проверки).
// Читает тело запроса, вычисляет ожидаемую подпись и сверяет с заголовком.
// После чтения восстанавливает тело запроса, чтобы его могли прочитать следующие middleware.
// Возвращает ошибку, если подпись не совпадает или заголовок имеет неверный формат.
func (s *Signer) CheckRequestSignature(r *http.Request) error {
	if r.Header.Get("HashSHA256") == "" {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return &ErrBodyRead{Source: "request", Method: r.Method, URL: r.URL.String(), BodySize: len(body), Err: err}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	headerValue := r.Header.Get("HashSHA256")
	received, err := hex.DecodeString(headerValue)
	if err != nil {
		return &ErrInvalidHexFormat{
			HeaderValue: headerValue,
			Err:         err,
		}
	}

	expected := s.CreateSignature(body)

	if !hmac.Equal(received, expected) {
		return &ErrSignedBodyMismatch{
			Expected: hex.EncodeToString(expected),
			Received: hex.EncodeToString(received),
		}
	}

	return nil
}
