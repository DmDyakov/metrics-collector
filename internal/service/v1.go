package service

import (
	"fmt"
	models "metrics-collector/internal/model"
	"strconv"
)

func (svc *MetricsService) UpdateMetric(metricType, metricName, metricValue string) error {
	input := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRequest, ErrInvalidCounterValue)
		}

		input.Delta = &delta

		err = svc.validateMetricFull(&input)
		if err != nil {
			return err
		}

		existing, ok := svc.repo.GetMetric(input.ID)
		if ok {
			if existing.MType != models.Counter {
				return fmt.Errorf("%w: %w", ErrInvalidRepoData, ErrMetricTypeMismatch)
			}
			if existing.Delta != nil {
				delta += *existing.Delta
			}
		}

		updatedMetric := svc.repo.UpdateMetric(models.Metrics{
			ID:    input.ID,
			MType: models.Counter,
			Delta: &delta,
		})

		if err = svc.validateMetricFull(updatedMetric); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
		}

	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRequest, ErrInvalidGaugeValue)
		}

		input.Value = &value

		updatedMetric := svc.repo.UpdateMetric(input)
		err = svc.validateMetricFull(updatedMetric)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
		}

	default:
		return fmt.Errorf("%w: %w", ErrInvalidRequest, ErrUnknownMetricType)
	}

	return nil
}

func (svc *MetricsService) GetAllMetrics() (map[string]string, error) {
	metrics := svc.repo.GetAllMetrics()

	allMetrics := make(map[string]string, len(metrics))

	for name, metric := range metrics {
		err := svc.validateMetricFull(&metric)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
		}
		value := formatToString(&metric)
		allMetrics[name] = value
	}

	return allMetrics, nil
}

func (svc *MetricsService) GetMetricValue(metricType, metricName string) (*string, error) {
	input := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	if err := svc.validateMetricBase(&input); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	m, ok := svc.repo.GetMetric(input.ID)
	if !ok {
		return nil, ErrMetricNotFound
	}

	if m.MType != input.MType {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, ErrMetricTypeMismatch)
	}

	err := svc.validateMetricFull(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
	}

	value := formatToString(m)

	return &value, nil
}

func formatToString(m *models.Metrics) string {
	var value string
	switch m.MType {
	case models.Counter:
		value = strconv.FormatInt(*m.Delta, 10)
	case models.Gauge:
		value = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	return value
}
