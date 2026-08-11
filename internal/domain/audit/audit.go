// Package audit defines the audit event type used throughout the application.
package audit

// Event represents a single audit log entry.
type Event struct {
	// Timestamp is the Unix timestamp of the audited action.
	Timestamp int64 `json:"ts"`

	// Metrics contains the metric names involved in the audited request.
	Metrics []string `json:"metrics"`

	// IPAddress is the remote IP address that made the request.
	IPAddress string `json:"ip_address"`
}
