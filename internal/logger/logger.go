// Package logger предоставляет функции для логирования с разными уровнями.
package logger

import (
	"go.uber.org/zap"
)

func NewZapLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
