package logger

import (
	"go.uber.org/zap"
)

func NewSugarZapLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()

	return sugar, nil
}
