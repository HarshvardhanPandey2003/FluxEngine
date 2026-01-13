package models

import (
	"time"
)

// Job represents a CPU-intensive unit of work for background processing.
// Each job contains metadata and a payload that determines what computation to perform.
type Job struct {
	ID        string                 `json:"id"`         // Unique identifier
	Type      string                 `json:"type"`       // "password_hash" or "report_generation"
	Payload   map[string]interface{} `json:"payload"`    // Job-specific data
	CreatedAt time.Time              `json:"created_at"` // Submission timestamp
	Status    string                 `json:"status"`     // "pending", "processing", "completed", "failed"
}

// JobResult holds the output of a completed job.
// This will be returned to demonstrate that real work was done.
type JobResult struct {
	JobID       string      `json:"job_id"`
	Duration    string      `json:"duration"`     // How long the job took
	Result      interface{} `json:"result"`       // The actual computation result
	CompletedAt time.Time   `json:"completed_at"`
}

// NewJob creates a new job instance with initialized fields.
func NewJob(jobType string, payload map[string]interface{}) *Job {
	return &Job{
		ID:        generateJobID(),
		Type:      jobType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Status:    "pending",
	}
}

// generateJobID creates a time-based unique identifier.
// Format: YYYYMMDDHHMMSS-randomsuffix
func generateJobID() string {
	return time.Now().Format("20060102150405") + "-" + randString(6)
}

// randString generates a random alphanumeric string.
// Uses current nanosecond timestamp for pseudo-randomness (simple but sufficient for IDs).
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		// Use nanoseconds for variety (not cryptographically secure, but fine for IDs)
		b[i] = letters[(time.Now().UnixNano()+int64(i))%int64(len(letters))]
	}
	return string(b)
}
