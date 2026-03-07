package repository

import (
	"testing"
)

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	storage.UpdateCounter("test", 10)

	val, exists := storage.counters["test"]
	if !exists {
		t.Error("counter 'test' not found")
	}
	if val != 10 {
		t.Errorf("got %d, want 10", val)
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	storage.UpdateGauge("test", 123.456)

	val, exists := storage.gauges["test"]
	if !exists {
		t.Error("gauge 'test' not found")
	}
	if val != 123.456 {
		t.Errorf("got %f, want 123.456", val)
	}
}
