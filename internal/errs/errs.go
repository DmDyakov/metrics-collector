package errs

import "errors"

var (
	ErrInvalidResponse = errors.New("invalid response data")
	ErrInvalidRequest  = errors.New("invalid request data")

	// metrics
	ErrMetricTypeRequired          = errors.New("metric type is required")
	ErrMetricNameRequired          = errors.New("metric name is required")
	ErrMetricValueForGaugeRequired = errors.New("metric value is required for gauge")
	ErrMetricDeltaForCountRequired = errors.New("metric delta is required for counter")
	ErrMetricTypeMismatch          = errors.New("metric type mismatch")
	ErrUnknownMetricType           = errors.New("unknown metric type")
	ErrMetricNotFound              = errors.New("metric not found")
	ErrInvalidCounterValue         = errors.New("invalid counter value, should be int")
	ErrInvalidGaugeValue           = errors.New("invalid gauge value, should be float")
)
