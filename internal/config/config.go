// Package config предоставляет функции для загрузки и парсинга конфигурации
// из переменных окружения и флагов командной строки.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func loadJSONFile(path string, v interface{}) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("could not parse config file %s: %w", path, err)
	}
	return nil
}

func loadDotEnv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not load .env file: %w", err)
	}
	return nil
}
