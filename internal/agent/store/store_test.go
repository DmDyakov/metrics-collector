package store

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_UpdateMetrics(t *testing.T) {
	const (
		cpuValue    = 42.0
		memoryValue = 1024.0
	)

	t.Run("adds metrics to store", func(t *testing.T) {
		s := New()

		metrics := map[string]float64{
			"cpu":    cpuValue,
			"memory": memoryValue,
		}

		s.UpdateMetrics(metrics)

		snapshot := s.GetMetricsSnapshot()
		assert.Equal(t, 2, len(snapshot))
		assert.Equal(t, cpuValue, snapshot["cpu"])
		assert.Equal(t, memoryValue, snapshot["memory"])
	})

	t.Run("overwrites existing metric", func(t *testing.T) {
		s := New()

		s.UpdateMetrics(map[string]float64{"cpu": 50.0})
		s.UpdateMetrics(map[string]float64{"cpu": cpuValue})

		snapshot := s.GetMetricsSnapshot()
		assert.Equal(t, 1, len(snapshot))
		assert.Equal(t, cpuValue, snapshot["cpu"])
	})
}

func TestStore_GetMetricsSnapshot(t *testing.T) {
	const cpuValue = 100.0

	t.Run("returns copy not reference", func(t *testing.T) {
		s := New()

		s.UpdateMetrics(map[string]float64{"cpu": cpuValue})

		snapshot := s.GetMetricsSnapshot()
		snapshot["cpu"] = 200.0

		actual := s.GetMetricsSnapshot()
		assert.Equal(t, cpuValue, actual["cpu"])
	})

	t.Run("empty store returns empty snapshot", func(t *testing.T) {
		s := New()

		snapshot := s.GetMetricsSnapshot()

		assert.NotNil(t, snapshot)
		assert.Equal(t, 0, len(snapshot))
	})
}

func TestStore_Concurrent(t *testing.T) {
	t.Run("concurrent read and write", func(t *testing.T) {
		s := New()
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				s.UpdateMetrics(map[string]float64{
					"metric": float64(i),
				})
			}(i)
		}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = s.GetMetricsSnapshot()
			}()
		}

		wg.Wait()
		// Не должно быть гонок (проверяется с -race)
	})
}

func TestStore_New(t *testing.T) {
	t.Run("creates store with initialized metrics", func(t *testing.T) {
		s := New()

		assert.NotNil(t, s)
		assert.NotNil(t, s.metrics)
	})
}
