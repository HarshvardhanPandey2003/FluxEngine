package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/HarshvardhanPandey2003/FluxEngine/internal/models"
)

// JobHandler manages HTTP endpoints for job submission.
type JobHandler struct {
	jobQueue chan<- *models.Job // Send-only channel
}

// NewJobHandler initializes the handler with a job queue reference.
func NewJobHandler(jobQueue chan<- *models.Job) *JobHandler {
	return &JobHandler{
		jobQueue: jobQueue,
	}
}

// SubmitJobRequest defines the expected JSON structure for job submissions.
type SubmitJobRequest struct {
	Type    string                 `json:"type"`    // "password_hash" or "report_generation"
	Payload map[string]interface{} `json:"payload"` // Job-specific parameters
}

// HandleSubmitJob processes POST requests to enqueue CPU-intensive background jobs.
func (h *JobHandler) HandleSubmitJob(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req SubmitJobRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate job type
	if req.Type != "password_hash" && req.Type != "report_generation" {
		http.Error(w, "Invalid job type. Use 'password_hash' or 'report_generation'", http.StatusBadRequest)
		return
	}

	// Validate required payload fields based on job type
	if req.Type == "password_hash" {
		if _, ok := req.Payload["password"]; !ok {
			http.Error(w, "Missing 'password' field in payload", http.StatusBadRequest)
			return
		}
	} else if req.Type == "report_generation" {
		if _, ok := req.Payload["data_points"]; !ok {
			// Default to 1 million data points if not specified
			req.Payload["data_points"] = 1000000
		}
	}

	// Create job instance
	job := models.NewJob(req.Type, req.Payload)

	// NON-BLOCKING SEND: Use select with default to avoid hanging
	// If the worker is busy and channel is full, we reject immediately
	select {
	case h.jobQueue <- job:
		// Job successfully queued
		log.Printf("✅ Job %s accepted (type: %s)", job.ID, job.Type)
	default:
		// Worker busy or queue full (in Phase 1, this shouldn't happen often)
		log.Printf("⚠️  Job queue full, rejecting job")
		http.Error(w, "System overloaded, try again later", http.StatusServiceUnavailable)
		return
	}

	// IMMEDIATE RESPONSE: Client doesn't wait for job completion
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 = Accepted but not yet processed

	response := map[string]interface{}{
		"status":  "accepted",
		"job_id":  job.ID,
		"message": fmt.Sprintf("Job %s queued for CPU-intensive processing", job.ID),
		"note":    "Job will be processed asynchronously. Use job_id to track status (future phases).",
	}

	json.NewEncoder(w).Encode(response)
}
